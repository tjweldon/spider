package messaging

import "tjweldon/spider/src/util"

type Backlog[T any] interface {
	Channel() <-chan T
	Length() int
}

type ForkedBacklog[T any] struct {
	output   <-chan T
	original Backlog[T]
}

func (fb *ForkedBacklog[T]) Channel() <-chan T {
	return fb.output
}

func (fb *ForkedBacklog[T]) Length() int {
	return fb.original.Length()
}

func Fork[T any](original Backlog[T]) (Backlog[T], Backlog[T]) {
	originalChannel := original.Channel()
	c1, c2 := util.Split[T](originalChannel, 0)
	return &ForkedBacklog[T]{c1, original}, &ForkedBacklog[T]{c2, original}
}
