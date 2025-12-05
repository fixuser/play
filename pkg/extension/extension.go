package extension

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

// Extension 是扩展需要实现的接口
type Extension interface {
	Name() string // Name 返回扩展的名称
	Load() error  // Load 加载扩展，如果失败则返回error
	Exit()        // Exit 退出扩展，不返回error，应确保资源释放
}

// Manager 用于管理扩展的生命周期
type Manager struct {
	mu                   sync.RWMutex
	registeredExtensions []Extension // 存储所有注册的扩展，保持注册顺序
	loadedExtensions     []Extension // 存储当前成功加载的扩展，保持加载顺序（与注册顺序一致）
}

// NewManager 创建一个新的 Manager 实例
func NewManager() *Manager {
	return &Manager{
		registeredExtensions: make([]Extension, 0),
		loadedExtensions:     make([]Extension, 0),
	}
}

// Register 用于注册一个扩展到管理器中
// 扩展将按照它们被注册的顺序进行加载
func (m *Manager) Register(ext Extension) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ext == nil {
		log.Warn().Msg("attempted to register a nil extension")
		return
	}
	m.registeredExtensions = append(m.registeredExtensions, ext)
	log.Trace().Str("extension_name", ext.Name()).Msg("extension registered")
}

// exitTheseExtensionsInReverse 是一个内部辅助函数，用于按反向顺序退出指定的扩展列表。
// 此函数由 LoadAll 在加载失败时（用于回滚）和 ExitAll（用于完全退出）调用。
func (m *Manager) exitTheseExtensionsInReverse(extensionsToExit []Extension) {
	if len(extensionsToExit) == 0 {
		log.Trace().Msg("no extensions provided to exit")
		return
	}
	log.Trace().Int("count", len(extensionsToExit)).Msg("starting to exit extensions in reverse order")
	// 从列表末尾向前遍历，实现反向退出
	for i := len(extensionsToExit) - 1; i >= 0; i-- {
		ext := extensionsToExit[i]
		log.Trace().Str("extension_name", ext.Name()).Msg("exiting extension")
		ext.Exit() // 调用扩展自身的退出逻辑
		log.Debug().Str("extension_name", ext.Name()).Msg("extension exited")
	}
	log.Trace().Msg("finished exiting extensions")
}

// LoadAll 尝试按照注册顺序加载所有已注册的扩展。
// 如果在加载过程中有任何扩展的 Load() 方法返回错误，
// 那么在此次 LoadAll 调用中已经成功加载的扩展将会被反向退出（回滚），
// 并且 LoadAll 会返回遇到的第一个错误。
// 只有当所有注册的扩展都成功加载后，manager内部的已加载列表才会更新。
func (m *Manager) LoadAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Trace().Msg("starting to load all registered extensions")

	if len(m.registeredExtensions) == 0 {
		log.Trace().Msg("no extensions registered, nothing to load")
		// m.loadedExtensions 应该已经是空的，或者保持上次成功加载的状态
		// 如果希望 LoadAll 清空之前的状态，可以在这里 m.loadedExtensions = make([]Extension, 0)
		// 但当前逻辑是，LoadAll 要么成功并用新的列表覆盖，要么失败并保持 m.loadedExtensions 不变
		return nil
	}

	// tempLoadedInThisCall 用于存储在本次 LoadAll 调用中成功加载的扩展。
	// 这样，如果中途发生失败，我们只退出在本次调用中已加载的扩展。
	var tempLoadedInThisCall []Extension

	for _, ext := range m.registeredExtensions {
		log.Trace().Str("extension_name", ext.Name()).Msg("attempting to load extension")
		if err := ext.Load(); err != nil {
			log.Error().Err(err).Str("extension_name", ext.Name()).Msg("failed to load extension")
			// 加载失败，反向退出在本次 LoadAll 调用中已经成功加载的扩展
			log.Trace().Msg("rolling back extensions loaded in this call due to error")
			m.exitTheseExtensionsInReverse(tempLoadedInThisCall) // 只退出本次调用中已加载的部分
			// m.loadedExtensions 保持不变 (即上一次成功加载的状态)
			return fmt.Errorf("failed to load extension '%s': %w", ext.Name(), err)
		}
		tempLoadedInThisCall = append(tempLoadedInThisCall, ext)
		log.Debug().Str("extension_name", ext.Name()).Msg("extension loaded successfully")
	}

	// 所有扩展都已成功加载
	m.loadedExtensions = tempLoadedInThisCall // 更新 Manager 的主加载列表
	log.Trace().Int("count", len(m.loadedExtensions)).Msg("all extensions loaded successfully")
	return nil
}

// ExitAll 负责反向退出所有当前已成功加载的扩展。
// “已经加载的扩展”指的是 manager 内部 `loadedExtensions` 列表中的扩展。
func (m *Manager) ExitAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Trace().Msg("starting to exit all currently loaded extensions")
	if len(m.loadedExtensions) == 0 {
		log.Trace().Msg("no extensions currently loaded, nothing to exit")
		return
	}
	m.exitTheseExtensionsInReverse(m.loadedExtensions)
	m.loadedExtensions = make([]Extension, 0) // 清空已加载列表，表示所有扩展均已退出
	log.Trace().Msg("all previously loaded extensions have been exited")
}
