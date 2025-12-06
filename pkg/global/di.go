package global

import (
	"sync/atomic"

	"github.com/play/play/pkg/di"
)

func defaultInjector() *atomic.Value {
	injector := di.New()
	v := &atomic.Value{}
	v.Store(injector)
	return v
}

var globalInjector = defaultInjector()

// SetInjector replaces the global injector with a new instance.
func SetInjector(injector *di.Injector) {
	if injector == nil {
		return
	}
	globalInjector.Store(injector)
}

// GetInjector returns the current global injector instance.
func GetInjector() *di.Injector {
	return globalInjector.Load().(*di.Injector)
}
