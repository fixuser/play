package tag

import (
	"fmt"
	"reflect"
)

func toInt64s(vs []any) (values []int64) {
	for _, v := range vs {
		val := reflect.ValueOf(v)
		if val.CanInt() {
			value := val.Int() - 1
			if value < 0 || value > 62 {
				panic(fmt.Errorf("tag value %d out of range [1, 63]", value+1))
			}
			values = append(values, val.Int()-1)
			continue
		}
	}
	return
}

// Tags 标签
type Tags int64

// Is 判断是否包含所有标签
func (tag Tags) Is(vs ...any) bool {
	for _, v := range toInt64s(vs) {
		if tag&(1<<v) == 0 {
			return false
		}
	}
	return true
}

// Has 判断是否包含任意一个标签
func (tag Tags) Has(vs ...any) bool {
	for _, v := range toInt64s(vs) {
		if tag&(1<<v) != 0 {
			return true
		}
	}
	return false
}

// Add 添加标签
func (tag *Tags) Add(vs ...any) *Tags {
	for _, v := range toInt64s(vs) {
		*tag |= 1 << v
	}
	return tag
}

// Clear 移除标签
func (tag *Tags) Clear(vs ...any) *Tags {
	for _, v := range toInt64s(vs) {
		*tag &^= 1 << v
	}
	return tag
}
