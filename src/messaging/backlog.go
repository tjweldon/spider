package messaging

import "tjweldon/spider/src/util"

// Backlog is the interface that the queue presents to a consumer.
type Backlog[T any] interface {
	Channel() <-chan T
	Length() int
}

// ForkedBacklog is a Backlog implementation that duplexes the messages
// so that they can be consumed by more than one consumer.
type ForkedBacklog[T any] struct {
	output   <-chan T
	original Backlog[T]
}

// Channel is the implementation of Backlog.Channel
func (fb *ForkedBacklog[T]) Channel() <-chan T {
	return fb.output
}

// Length proxies back to the parent backlog
func (fb *ForkedBacklog[T]) Length() int {
	return fb.original.Length()
}

// Fork takes a Backlog and returns a pair of backlogs. The zero buffer size
// means that the messages have to be consumed at the same rate
func Fork[T any](original Backlog[T]) (Backlog[T], Backlog[T]) {
	originalChannel := original.Channel()
	c1, c2 := util.Split[T](originalChannel, 0)
	return &ForkedBacklog[T]{c1, original}, &ForkedBacklog[T]{c2, original}
}
