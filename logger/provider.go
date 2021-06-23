package logger

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

// Provider define an interface for writing logs
type Provider interface {
	Write(level Level, headerLength int, data []byte) error
	Close() error
}

// ProviderCreator is a factory function type for creating Provider
type ProviderCreator func(opts string) Provider

var (
	providersMu sync.Mutex
	providers   = map[string]ProviderCreator{}
)

// Register registers a provider by name and creator
func Register(providerType string, creator ProviderCreator) {
	providersMu.Lock()
	defer providersMu.Unlock()
	if _, ok := providers[providerType]; ok {
		panic("provider " + providerType + " registered")
	}
	providers[providerType] = creator
}

// Lookup gets provider creator by name
func Lookup(providerType string) ProviderCreator {
	providersMu.Lock()
	defer providersMu.Unlock()
	return providers[providerType]
}

// UnmarshalOpts unmarshal JSON string opts to object v
func UnmarshalOpts(opts string, v interface{}) error {
	opts = strings.TrimSpace(opts)
	if opts == "" {
		return nil
	}
	if opts[0] != '[' && opts[0] != '{' {
		jsonString, err := form2JSON(opts)
		if err != nil {
			return err
		}
		opts = jsonString
	}
	return json.Unmarshal([]byte(opts), v)
}

func form2JSON(form string) (jsonString string, err error) {
	m := map[string]interface{}{}
	for form != "" {
		key := form
		if i := strings.IndexAny(key, "&;"); i >= 0 {
			key, form = key[:i], key[i+1:]
		} else {
			form = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = url.QueryUnescape(value)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		m[key] = parseValue(value)
	}
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parseValue(s string) interface{} {
	if v, err := strconv.ParseInt(s, 10, 64); err == nil {
		return v
	}
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}
	if v, err := strconv.ParseBool(s); err == nil {
		return v
	}
	return s
}
