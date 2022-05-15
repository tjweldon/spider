package messaging

import "tjweldon/spider/src/util"

type Backlog[T any] interface {
	Channel() <-chan T
	Length() int
}

type ForkedBacklog[T any] struct {
	output  <-chan T
	orginal Backlog[T]
}

func (fb *ForkedBacklog[T]) Channel() <-chan T {
	return fb.output
}

func (fb *ForkedBacklog[T]) Length() int {
	return fb.orginal.Length()
}

func Fork[T any](b Backlog[T]) (Backlog[T], Backlog[T]) {
	original := b.Channel()
	c1, c2 := util.Split[T](original, 0)
	return &ForkedBacklog[T]{c1, b}, &ForkedBacklog[T]{c2, b}
}
