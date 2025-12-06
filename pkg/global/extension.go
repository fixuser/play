package global

import (
	"sync/atomic"

	"github.com/play/play/pkg/extension"
)

func defaultExtensionManager() *atomic.Value {
	v := &atomic.Value{}
	v.Store(extension.NewManager())
	return v
}

var globalExtensionManager = defaultExtensionManager()

// SetExtensionManager sets the global extension manager.
func SetExtensionManager(m *extension.Manager) {
	globalExtensionManager.Store(m)
}

// GetExtensionManager retrieves the current global extension manager.
func GetExtensionManager() *extension.Manager {
	return globalExtensionManager.Load().(*extension.Manager)
}
