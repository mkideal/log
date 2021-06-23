package log

import (
	"bytes"
)

type entry struct {
	bytes.Buffer
	tmp    [64]byte
	next   *entry
	level  Level
	header int
}

func (e *entry) Reset() {
	e.Buffer.Reset()
	e.header = 0
}

func (e *entry) Level() Level { return e.level }

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
