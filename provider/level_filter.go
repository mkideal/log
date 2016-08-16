package provider

import (
	"github.com/mkideal/log/logger"
)

type LevelFilterFunc func(level logger.Level) bool

func IsLevel(level logger.Level) LevelFilterFunc {
	return func(lv logger.Level) bool { return level == lv }
}

type LevelFilter struct {
	provider logger.Provider
	filter   LevelFilterFunc
}

func NewLevelFilter(provider logger.Provider, filter LevelFilterFunc) logger.Provider {
	return &LevelFilter{
		provider: provider,
		filter:   filter,
	}
}

func (p *LevelFilter) Write(level logger.Level, headerLength int, data []byte) error {
	if p.filter(level) {
		return p.provider.Write(level, headerLength, data)
	}
	return nil
}

func (p *LevelFilter) Close() error { return p.provider.Close() }
