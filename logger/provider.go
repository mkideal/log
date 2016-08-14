package logger

import (
	"encoding/json"
)

type Provider interface {
	Write(level Level, headerLength int, data []byte) error
	Close() error
}

type ProviderCreator func(opts string) Provider

var (
	providers = map[string]ProviderCreator{}
)

func Register(providerType string, creator ProviderCreator) {
	if _, ok := providers[providerType]; ok {
		panic("provider " + providerType + " registered")
	}
	providers[providerType] = creator
}

func Lookup(providerType string) ProviderCreator {
	return providers[providerType]
}

func UnmarshalOpts(opts string, v interface{}) error {
	if opts == "" {
		return nil
	}
	return json.Unmarshal([]byte(opts), v)
}
