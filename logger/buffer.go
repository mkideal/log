package logger

import (
	"bytes"
)

type buffer struct {
	bytes.Buffer
	tmp          [32]byte
	next         *buffer
	level        Level
	headerLength int
	quit         bool
}

const digits = "0123456789"

func twoDigits(buf *buffer, begin int, v int) {
	buf.tmp[begin+1] = digits[v%10]
	v /= 10
	buf.tmp[begin] = digits[v%10]
}

func threeDigits(buf *buffer, begin int, v int) {
	buf.tmp[begin+2] = digits[v%10]
	v /= 10
	buf.tmp[begin+1] = digits[v%10]
	v /= 10
	buf.tmp[begin] = digits[v%10]
}

func fourDigits(buf *buffer, begin int, v int) {
	buf.tmp[begin+3] = digits[v%10]
	v /= 10
	buf.tmp[begin+2] = digits[v%10]
	v /= 10
	buf.tmp[begin+1] = digits[v%10]
	v /= 10
	buf.tmp[begin] = digits[v%10]
}

func someDigits(buf *buffer, begin int, v int) int {
	j := len(buf.tmp)
	for {
		j--
		buf.tmp[j] = digits[v%10]
		v /= 10
		if v == 0 {
			break
		}
	}
	return copy(buf.tmp[begin:], buf.tmp[j:])
}
