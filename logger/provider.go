package logger

import (
	"encoding/json"
)

// Provider define an interface for writting logs
type Provider interface {
	Write(level Level, headerLength int, data []byte) error
	Close() error
}

// ProviderCreator is a factory function type for creating Provider
type ProviderCreator func(opts string) Provider

var (
	providers = map[string]ProviderCreator{}
)

// Register registers a provider by name and creator
func Register(providerType string, creator ProviderCreator) {
	if _, ok := providers[providerType]; ok {
		panic("provider " + providerType + " registered")
	}
	providers[providerType] = creator
}

// Lookup gets provider creator by name
func Lookup(providerType string) ProviderCreator {
	return providers[providerType]
}

// UnmarshalOpts unmarshal JSON string opts to object v
func UnmarshalOpts(opts string, v interface{}) error {
	if opts == "" {
		return nil
	}
	return json.Unmarshal([]byte(opts), v)
}
