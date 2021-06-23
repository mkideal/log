package log

import (
	"os"
	"testing"

	"github.com/mkideal/log/provider"
	"github.com/stretchr/testify/assert"
)

func TestParseLevel(t *testing.T) {
	lv, ok := ParseLevel("trace")
	assert.Equal(t, lv, LvTRACE)
	assert.True(t, ok)

	lv, ok = ParseLevel("debug")
	assert.Equal(t, lv, LvDEBUG)
	assert.True(t, ok)

	lv, ok = ParseLevel("info")
	assert.Equal(t, lv, LvINFO)
	assert.True(t, ok)

	lv, ok = ParseLevel("warn")
	assert.Equal(t, lv, LvWARN)
	assert.True(t, ok)

	lv, ok = ParseLevel("error")
	assert.Equal(t, lv, LvERROR)
	assert.True(t, ok)

	lv, ok = ParseLevel("fatal")
	assert.Equal(t, lv, LvFATAL)
	assert.True(t, ok)

	lv, ok = ParseLevel("invalid")
	assert.Equal(t, lv, LvINFO)
	assert.False(t, ok)

	assert.Equal(t, MustParseLevel("trace"), LvTRACE)
	assert.Equal(t, MustParseLevel("debug"), LvDEBUG)
	assert.Equal(t, MustParseLevel("info"), LvINFO)
	assert.Equal(t, MustParseLevel("warn"), LvWARN)
	assert.Equal(t, MustParseLevel("error"), LvERROR)
	assert.Equal(t, MustParseLevel("fatal"), LvFATAL)
	assert.Panics(t, func() { MustParseLevel("invalid") })
}

func TestInit(t *testing.T) {
	defer func() {
		os.RemoveAll("./log")
	}()
	assert.Error(t, Init("", ""))
	assert.Error(t, Init("not-found-provider", ""))

	assert.Nil(t, Init("console", ""))
	assert.Nil(t, Init("console/file", `{"dir": "./log"}`))

	Uninit(InitConsole(LvWARN))
	Uninit(InitFile("./log/app.log"))
	Uninit(InitFileAndConsole("./log/app.log", LvWARN))
	Uninit(InitMultiFile("./log", "app.log"))
	Uninit(InitMultiFileAndConsole("./log", "app.log", LvWARN))
	Uninit(InitSyncWithProvider(provider.NewConsole("")))
}

func TestSetLevelFromString(t *testing.T) {
	assert.Equal(t, LvTRACE, SetLevelFromString("trace"))
	assert.Equal(t, LvTRACE, GetLevel())
	assert.Equal(t, LvDEBUG, SetLevelFromString("debug"))
	assert.Equal(t, LvDEBUG, GetLevel())
	assert.Equal(t, LvINFO, SetLevelFromString("info"))
	assert.Equal(t, LvINFO, GetLevel())
	assert.Equal(t, LvWARN, SetLevelFromString("warn"))
	assert.Equal(t, LvWARN, GetLevel())
	assert.Equal(t, LvERROR, SetLevelFromString("error"))
	assert.Equal(t, LvERROR, GetLevel())
	assert.Equal(t, LvFATAL, SetLevelFromString("fatal"))
	assert.Equal(t, LvFATAL, GetLevel())
	assert.Equal(t, LvINFO, SetLevelFromString("invalid"))
	assert.Equal(t, LvINFO, GetLevel())
}
