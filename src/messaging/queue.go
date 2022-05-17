package messaging

import (
	"log"
	"tjweldon/spider/src/util"
)

// Queue is the type underlying all the dispatchers and backlogs.
// It implements the buffered channel that is the actual message queue.
type Queue[T any] struct {
	input  chan T
	output chan T
	queue  chan T
}

// NewQ constructs a Queue with a buffer of the passed size. It returns
// this as a Dispatcher and a Backlog pair to be passed to different
// processes.
func NewQ[T any](size int) (Dispatcher[T], Backlog[T]) {
	q := &Queue[T]{
		input:  make(chan T),
		output: make(chan T),
		queue:  make(chan T, size),
	}
	return q.Split()
}

// Split is the the method that starts the send and receive generators
// that put on to and take from the Queue respectively
func (q *Queue[T]) Split() (Dispatcher[T], Backlog[T]) {
	receive := func(incoming <-chan T, queue chan<- T) {
		defer close(queue)
		for msg := range incoming {
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

// Channel returns the read side generator channel
func (q *Queue[T]) Channel() <-chan T {
	return q.output
}

// Dispatch sends messages into the write side of the send channel
func (q *Queue[T]) Dispatch(item T) (ok bool) {
	log.Printf("Dispatching: %v\n", item)
	q.input <- item
	return true
}

// Close closes the input channel which cascades through the send generator, to the
// queue and subsequently any consumers of the queue
func (q *Queue[T]) Close() {
	if !util.IsClosed[T](q.input) {
		close(q.input)
	}
}

// Length returns the number of messages currently in the queue buffer.
func (q *Queue[T]) Length() int {
	return len(q.queue)
}
