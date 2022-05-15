package messaging

type Dispatcher[T any] interface {
	Dispatch(item T) (ok bool)
	Close()
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
		maxJobs:       0,
	}
}

func (dd *DeDuplicatingDispatcher[T]) SetMaxJobs(max int) *DeDuplicatingDispatcher[T] {
	dd.maxJobs = max
	return dd
}

func (dd *DeDuplicatingDispatcher[T]) Dispatch(item T) bool {
	// Deduplication, ignores messages that have already been sent
	for _, prevItem := range dd.previousItems {
		if prevItem == item {
			return true
		}
	}

	// If the dispatcher has max jobs set, and we have done
	// more than the max jobs, close the dispatcher and return.
	if dd.maxJobs > 0 && len(dd.previousItems) >= dd.maxJobs {
		return false
	}

	// Record the new unique message to prevent it being
	// sent again.
	dd.previousItems = append(dd.previousItems, item)

	// Dispatch the message
	return dd.dispatcher.Dispatch(item)
}

func (dd *DeDuplicatingDispatcher[T]) Close() {
	dd.dispatcher.Close()
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

func (vd *ValidDispatcher[T]) Dispatch(item T) (ok bool) {
	for _, validator := range vd.validators {
		if !validator(item) {
			return true
		}
	}

	return vd.dispatcher.Dispatch(item)
}

func (vd *ValidDispatcher[T]) Close() {
	vd.dispatcher.Close()
}

type PreProcessor[T any] func(item T) T

type PreProcessingDispatcher[T any] struct {
	dispatcher    Dispatcher[T]
	preProcessors []PreProcessor[T]
}

func WithPreProcessing[T any](
	dispatcher Dispatcher[T], preProcessors ...PreProcessor[T],
) *PreProcessingDispatcher[T] {
	return &PreProcessingDispatcher[T]{
		dispatcher:    dispatcher,
		preProcessors: preProcessors,
	}
}

func (ppd *PreProcessingDispatcher[T]) Dispatch(item T) (ok bool) {
	for _, preProcessor := range ppd.preProcessors {
		item = preProcessor(item)
	}

	return ppd.dispatcher.Dispatch(item)
}

func (ppd *PreProcessingDispatcher[T]) Close() {
	ppd.dispatcher.Close()
}
