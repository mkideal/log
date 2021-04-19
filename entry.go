package log

import (
	"bytes"
)

type entry struct {
	bytes.Buffer
	tmp       [32]byte
	next      *entry
	level     Level
	headerLen int
	timestamp int64
}

func (e *entry) Reset() {
	e.Buffer.Reset()
	e.headerLen = 0
}

func (e *entry) clone() *entry {
	e2 := &entry{
		level:     e.level,
		headerLen: e.headerLen,
		timestamp: e.timestamp,
	}
	e2.Buffer = bytes.Buffer{}
	e2.Buffer.Write(e.Bytes())
	return e2
}

func (e *entry) Level() Level     { return e.level }
func (e *entry) Timestamp() int64 { return e.timestamp }

const digits = "0123456789"

func twoDigits(e *entry, begin int, v int) {
	e.tmp[begin+1] = digits[v%10]
	v /= 10
	e.tmp[begin] = digits[v%10]
}

func threeDigits(e *entry, begin int, v int) {
	e.tmp[begin+2] = digits[v%10]
	v /= 10
	e.tmp[begin+1] = digits[v%10]
	v /= 10
	e.tmp[begin] = digits[v%10]
}

func fourDigits(e *entry, begin int, v int) {
	e.tmp[begin+3] = digits[v%10]
	v /= 10
	e.tmp[begin+2] = digits[v%10]
	v /= 10
	e.tmp[begin+1] = digits[v%10]
	v /= 10
	e.tmp[begin] = digits[v%10]
}

func someDigits(e *entry, begin int, v int) int {
	j := len(e.tmp)
	for {
		j--
		e.tmp[j] = digits[v%10]
		v /= 10
		if v == 0 {
			break
		}
	}
	return copy(e.tmp[begin:], e.tmp[j:])
}

// entry queue
type queue struct {
	in  []*entry
	out []*entry
}

func newQueue() *queue {
	return &queue{
		in: make([]*entry, 0, 64),
	}
}

func (q *queue) size() int {
	return len(q.in)
}

func (q *queue) push(e *entry) int {
	q.in = append(q.in, e)
	return len(q.in)
}

func (q *queue) popAll() []*entry {
	q.in, q.out = q.out, q.in
	q.in = q.in[:0]
	return q.out
}
