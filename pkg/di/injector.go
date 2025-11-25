package di

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/rs/zerolog/log"
)

// Injector 是一个依赖注入容器，负责存储和解析依赖项。
type Injector struct {
	mu           sync.RWMutex         // mu 用于确保对 dependencies 映射的并发安全访问。
	dependencies map[reflect.Type]any // dependencies 存储已注册的依赖项实例，键是其反射类型。
}

// New 创建一个新的 Injector 实例
func New() *Injector {
	return &Injector{
		mu:           sync.RWMutex{},
		dependencies: make(map[reflect.Type]any),
	}
}

// Set (容器方法) 将一个或多个依赖项实例注册到当前容器中。
// vs: 一个或多个要注册的依赖项实例。
//
//	建议注册指针类型的服务实例，例如 &UserService{}，以便共享和通过接口解析。
func (c *Injector) Set(vs ...any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, v := range vs {
		if v == nil {
			log.Warn().Msg("attempted to set a nil dependency, skipping")
			continue
		}
		valType := reflect.TypeOf(v)
		c.dependencies[valType] = v
		log.Trace().Str("type", valType.String()).Msg("dependency set")
	}
}

// resolve (容器内部方法) 尝试从容器中解析给定类型的依赖项。
// requestedType: 需要解析的依赖项的 reflect.Type。
// 返回:
//   - reflect.Value: 解析到的依赖项的反射值。
//   - bool: 如果找到并成功解析了依赖项，则为 true；否则为 false。
func (c *Injector) resolve(requestedType reflect.Type) (reflect.Value, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 1. 直接类型匹配
	// 检查容器中是否存在与 requestedType 完全相同的类型。
	if instance, ok := c.dependencies[requestedType]; ok {
		log.Trace().Str("type", requestedType.String()).Msg("direct type match found in injector")
		return reflect.ValueOf(instance), true
	}

	// 2. 接口实现匹配
	// 如果 requestedType 是一个接口类型。
	if requestedType.Kind() == reflect.Interface {
		log.Trace().Str("interface_type", requestedType.String()).Msg("looking for implementations for requested interface")
		// 遍历容器中所有已注册的依赖项。
		for storedKeyType, storedInstance := range c.dependencies {
			instanceVal := reflect.ValueOf(storedInstance)
			// 检查存储的实例类型是否实现了请求的接口类型。
			// instanceVal.Type() 是存储实例的实际类型 (例如 *SomeStruct 或 SomeStruct)。
			if instanceVal.IsValid() && instanceVal.Type().Implements(requestedType) {
				log.Trace().
					Str("interface_type", requestedType.String()).
					Str("implementation_type", instanceVal.Type().String()).
					Str("stored_key_type", storedKeyType.String()). // 记录容器中医该实例的键的类型
					Msg("interface implementation found")
				return instanceVal, true
			}
		}
	}

	// 3. 自动解引用指针：如果请求的是 T (非指针)，但容器中存储的是 *T。
	// 例如：Get 时传入 var s MyStruct; (调用 Get(&s))，此时 requestedType 为 MyStruct。
	// 而 Set 时传入的是 &MyService{}，容器中存储的键为 *MyStruct。
	if requestedType.Kind() != reflect.Ptr {
		// 构造指向 requestedType 的指针类型，即 *T。
		ptrToRequestedType := reflect.PointerTo(requestedType)
		if instance, ok := c.dependencies[ptrToRequestedType]; ok {
			instanceVal := reflect.ValueOf(instance) // instanceVal 的类型是 *T (指针)
			// 确保 instanceVal 是有效的、非nil的指针，并且其元素类型与 requestedType 匹配。
			if instanceVal.IsValid() && instanceVal.Kind() == reflect.Ptr && !instanceVal.IsNil() && instanceVal.Elem().Type() == requestedType {
				log.Trace().
					Str("requested_type", requestedType.String()).
					Str("found_pointer_type", ptrToRequestedType.String()).
					Msg("found pointer for requested non-pointer type, returning element")
				return instanceVal.Elem(), true // 返回指针指向的元素 T
			}
		}
	}
	// 注意：目前不支持相反情况的自动处理（例如，请求 *T，但容器中存储的是 T）。
	// 这种从值到指针的转换更为复杂，因为map中的值不是直接可寻址的，需要创建新指针，
	// 这可能不是用户期望的行为（用户可能期望共享的是原始注册的实例）。
	// 建议用户在 Set 时注册他们期望 Get 的确切类型，或者主要注册指针类型以便于共享和接口解析。

	log.Trace().Str("type", requestedType.String()).Msg("no suitable dependency found after all checks")
	return reflect.Value{}, false
}

// Get (容器方法) 从容器中检索依赖项，并将它们注入到提供的目标变量中。
// targets: 一个或多个指向目标变量的指针。依赖项将被注入到这些变量中。
// 返回: 如果任何依赖项无法找到、类型不匹配或无法设置，则返回错误。
func (c *Injector) Get(targets ...any) error {
	for i, target := range targets {
		targetVal := reflect.ValueOf(target)

		// 验证 target 是否为指针类型
		if targetVal.Kind() != reflect.Ptr {
			err := fmt.Errorf("target at index %d is not a pointer (type: %s)", i, targetVal.Type())
			log.Error().Err(err).Int("index", i).Str("type", targetVal.Type().String()).Msg("get target must be a pointer")
			return err
		}

		// 验证 target 指针是否为 nil
		if targetVal.IsNil() {
			err := fmt.Errorf("target pointer at index %d is nil (type: %s)", i, targetVal.Type())
			log.Error().Err(err).Int("index", i).Str("type", targetVal.Type().String()).Msg("get target cannot be a nil pointer")
			return err
		}

		// elemToSet 是指针指向的实际变量值，我们需要对其进行设置。
		elemToSet := targetVal.Elem()
		// typeToFind 是实际变量的类型，我们将在容器中查找此类型的依赖项。
		typeToFind := elemToSet.Type()

		log.Trace().Str("target_type", typeToFind.String()).Int("index", i).Msg("attempting to get dependency")

		// 从容器中解析依赖项
		resolvedDepVal, found := c.resolve(typeToFind)
		if !found {
			err := fmt.Errorf("dependency not found for type '%s' for target at index %d", typeToFind.String(), i)
			log.Warn().Str("type", typeToFind.String()).Int("index", i).Msg("dependency not found in injector")
			return err // 如果一个依赖项找不到，则整体 Get 操作失败
		}

		// 检查解析到的依赖项是否可以赋值给目标变量
		// resolvedDepVal.Type() 是实际找到的依赖项的类型
		// elemToSet.Type() 是目标变量的类型
		if !resolvedDepVal.Type().AssignableTo(elemToSet.Type()) {
			err := fmt.Errorf("resolved dependency of type '%s' is not assignable to target type '%s' for target at index %d",
				resolvedDepVal.Type().String(), elemToSet.Type().String(), i)
			log.Error().Err(err).
				Int("index", i).
				Str("resolved_type", resolvedDepVal.Type().String()).
				Str("target_type", elemToSet.Type().String()).
				Msg("type mismatch prevents assignment")
			return err
		}

		// 检查目标变量是否可以被设置
		if !elemToSet.CanSet() {
			err := fmt.Errorf("cannot set value for target type '%s' at index %d (is it an unexported field or an unaddressable value?)", elemToSet.Type().String(), i)
			log.Error().Err(err).Int("index", i).Str("type", elemToSet.Type().String()).Msg("cannot set target variable")
			return err
		}

		// 设置依赖项到目标变量
		elemToSet.Set(resolvedDepVal)
		log.Trace().Str("type", typeToFind.String()).Int("index", i).Msg("dependency retrieved and set successfully")
	}
	return nil
}

// MustGet (容器方法) 从容器中检索依赖项，并将它们注入到提供的目标变量中。
func (c *Injector) MustGet(targets ...any) {
	err := c.Get(targets...)
	if err != nil {
		panic(err)
	}
}
