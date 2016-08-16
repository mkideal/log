package main

import (
	"github.com/mkideal/log"
)

func main() {
	defer log.Uninit(log.InitConsole(log.LvERROR))

	log.With("hello,with").Info("with a string")
	log.With([]byte("hello,with")).Info("with a bytes")
	log.With(0).Info("with an int")
	log.With(int8(8)).Info("with an int8")
	log.With(int16(16)).Info("with an int16")
	log.With(int32(32)).Info("with an int32")
	log.With(int64(64)).Info("with an int64")
	log.With(uint(0)).Info("with an uint")
	log.With(uintptr(0xff)).Info("with an uintptr")
	log.With(uint8(80)).Info("with an uint8")
	log.With(uint16(160)).Info("with an uint16")
	log.With(uint32(320)).Info("with an uint32")
	log.With(uint64(640)).Info("with an uint64")
	log.With(true).Info("with a true")
	log.With(false).Info("with a false")
	log.With(log.M{"a": 1, "b": "haha"}).Info("with a map")
	log.With(log.S{"a", 1}).Info("with a slice")
	log.WithJSON(log.M{"a": 1, "b": "haha"}).Info("with a map json")
	log.WithJSON(log.S{"a", 1}).Info("with a slice json")

	log.With(log.M{"a": 1}).With(log.M{"b": true}).Info("with called 2 times: M and M")
	log.With(log.S{"a"}).With(log.S{true}).Info("with called 2 times: S and S")
	log.With(log.M{"a": 1}).With(log.S{true}).Info("with called 2 times: M and S")
	log.With(log.S{"a"}).With(log.M{"b": true}).Info("with called 2 times: S and M")
	log.With(log.S{true, false}).With(2).Info("with called 2 times: S and object")
	log.With(2).With(log.S{true, false}).Info("with called 2 times: object and S")
	log.With(2).With(true).Info("with called 2 times: object and object")
	log.With(log.S{"a"}).With(true).With(false).Info("with called 3 times: S and object and object")

	log.If(true).With("haha").Info("should be printed")
	log.If(false).With("haha").Info("shouldn't be printed")
}
