package messaging

import (
	"log"
	"tjweldon/spider/src/util"
)

type Queue[T any] struct {
	input  chan T
	output chan T
	queue  chan T
}

func NewQ[T any](size int) (Dispatcher[T], Backlog[T]) {
	q := &Queue[T]{
		input:  make(chan T),
		output: make(chan T),
		queue:  make(chan T, size),
	}
	return q.Split()
}

func (q *Queue[T]) Split() (Dispatcher[T], Backlog[T]) {
	receive := func(in <-chan T, queue chan<- T) {
		defer close(queue)
		for msg := range in {
			queue <- msg
		}
	}

	deliver := func(out chan<- T, queue <-chan T) {
		defer close(out)
		for msg := range queue {
			out <- msg
		}
	}

	go receive(q.input, q.queue)
	go deliver(q.output, q.queue)

	return q, q
}

func (q *Queue[T]) Channel() <-chan T {
	return q.output
}

func (q *Queue[T]) Dispatch(item T) (ok bool) {
	log.Printf("Dispatching: %v\n", item)
	q.input <- item
	return true
}

func (q *Queue[T]) Close() {
	if !util.IsClosed[T](q.input) {
		close(q.input)
	}
}

func (q *Queue[T]) Length() int {
	return len(q.queue)
}
