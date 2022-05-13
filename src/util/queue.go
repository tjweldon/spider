package util

import (
	"log"
	"time"
)

type Dispatcher[T any] interface {
	Dispatch(item T)
}

type DeDuplicatingDispatcher[T comparable] struct {
	dispatcher    Dispatcher[T]
	previousItems []T
	maxJobs       int
}

func WithDeDuplication[T comparable](dispatcher Dispatcher[T]) *DeDuplicatingDispatcher[T] {
	return &DeDuplicatingDispatcher[T]{
		dispatcher:    dispatcher,
		previousItems: []T{},
		maxJobs:       -1,
	}
}

func (dd *DeDuplicatingDispatcher[T]) SetMaxJobs(max int) *DeDuplicatingDispatcher[T] {
	dd.maxJobs = max
	return dd
}

func (dd *DeDuplicatingDispatcher[T]) Dispatch(item T) {
	for _, prevItem := range dd.previousItems {
		if prevItem == item {
			return
		}
	}
	if dd.maxJobs >= 0 && len(dd.previousItems) >= dd.maxJobs {
		return
	}
	dd.previousItems = append(dd.previousItems, item)
	dd.dispatcher.Dispatch(item)
}

func (dd *DeDuplicatingDispatcher[T]) ReportDispatched() []T {
	return dd.previousItems
}

type Validator[T any] func(job T) bool

type ValidDispatcher[T any] struct {
	dispatcher Dispatcher[T]
	validators []Validator[T]
}

func WithValidation[T any](dispatcher Dispatcher[T], validators ...Validator[T]) *ValidDispatcher[T] {
	return &ValidDispatcher[T]{
		dispatcher: dispatcher,
		validators: validators,
	}
}

func (vd *ValidDispatcher[T]) Dispatch(item T) {
	for _, validator := range vd.validators {
		if !validator(item) {
			return
		}
	}

	vd.dispatcher.Dispatch(item)
}

type PreProcessor[T any] func(item T) T

type PreProcessingDispatcher[T any] struct {
	dispatcher    Dispatcher[T]
	preProcessors []PreProcessor[T]
}

func WithPreProcessing[T any](
	dispatcher Dispatcher[T],
	preProcessors ...PreProcessor[T],
) *PreProcessingDispatcher[T] {
	return &PreProcessingDispatcher[T]{
		dispatcher:    dispatcher,
		preProcessors: preProcessors,
	}
}

func (pv *PreProcessingDispatcher[T]) Dispatch(item T) {
	for _, preProcessor := range pv.preProcessors {
		item = preProcessor(item)
	}

	pv.dispatcher.Dispatch(item)
}

type Backlog[T any] interface {
	TakeOne() (item T, ok bool)
	Channel() <-chan T
}

type Queue[T any] struct {
	messages *Deque[T]
	input    chan T
	output   chan T
}

func NewQ[T any]() (Dispatcher[T], Backlog[T]) {
	return (&Queue[T]{
		messages: NewDeque[T](),
		input:    make(chan T),
		output:   make(chan T),
	}).Split()
}

func (q *Queue[T]) Split() (Dispatcher[T], Backlog[T]) {
	messages := q.messages

	receive := func(in <-chan T) {
		for msg := range in {
			messages.Insert(msg)
		}
	}

	deliver := func(out chan<- T) {
		for {
			if messages.IsEmpty() {
				time.Sleep(time.Second / 10)
				continue
			}
			out <- messages.TakeOne()
		}
	}

	go receive(q.input)
	go deliver(q.output)

	return q, q
}

func (q *Queue[T]) TakeOne() (item T, ok bool) {
	select {
	case item = <-q.output:
		return item, true
	default:
		return
	}
}

func (q *Queue[T]) Channel() <-chan T {
	return q.output
}

func (q *Queue[T]) Dispatch(item T) {
	log.Printf("Dispatching: %v\n", item)
	q.input <- item
}
