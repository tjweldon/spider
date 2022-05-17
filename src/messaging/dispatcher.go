package messaging

// Dispatcher is the interface that the queue presents
// to whichever process wants to send messages.
type Dispatcher[T any] interface {

	// Dispatch is the function called by the sending process to put
	// messages onto the queue.
	Dispatch(item T) (ok bool)

	// Close indicates to downstream consumers that no more messages
	// will be forthcoming
	Close()
}

// DeDuplicatingDispatcher is a Dispatcher implementation that
// will silently ignore messages with identical content
type DeDuplicatingDispatcher[T comparable] struct {
	dispatcher    Dispatcher[T]
	previousItems []T
	maxJobs       int
}

// WithDeDuplication wraps a dispatcher with a DeDuplicatingDispatcher
func WithDeDuplication[T comparable](dispatcher Dispatcher[T]) *DeDuplicatingDispatcher[T] {
	return &DeDuplicatingDispatcher[T]{
		dispatcher:    dispatcher,
		previousItems: []T{},
		maxJobs:       0,
	}
}

// SetMaxJobs is a fluent setter for the maximum number of unique URls after
// which all messages are ignored.
func (dd *DeDuplicatingDispatcher[T]) SetMaxJobs(max int) *DeDuplicatingDispatcher[T] {
	dd.maxJobs = max
	return dd
}

// Dispatch implements the deduplication and job limit.
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

// Close is just a proxy for everything but the underlying queue
func (dd *DeDuplicatingDispatcher[T]) Close() {
	dd.dispatcher.Close()
}

// ReportDispatched returns a slice of all of the items sent
// by the deduplicating dispatcher
func (dd *DeDuplicatingDispatcher[T]) ReportDispatched() []T {
	return dd.previousItems
}

// Validator is a function that acts as a filter for jobs
type Validator[T any] func(job T) bool

// ValidDispatcher is a configurable dispatcher that allows the
// client code to supply the Validator filtering functions.
type ValidDispatcher[T any] struct {
	dispatcher Dispatcher[T]
	validators []Validator[T]
}

// WithValidation wraps the passed dispatcher with validation filters
func WithValidation[T any](dispatcher Dispatcher[T], validators ...Validator[T]) *ValidDispatcher[T] {
	return &ValidDispatcher[T]{
		dispatcher: dispatcher,
		validators: validators,
	}
}

// Dispatch just iterates over Validators and returns ok=true the first time
// validation fails. If it doesn't fail, it calls its internal Dispatcher's
// Dispatch method
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

// PreProcessor is a method signature for a type invariant transformation
type PreProcessor[T any] func(item T) T

// PreProcessingDispatcher is a configurable Dispatcher that applies a series of
// transformations to a job, such as prepending relative paths with a hostname.
type PreProcessingDispatcher[T any] struct {
	dispatcher    Dispatcher[T]
	preProcessors []PreProcessor[T]
}

// WithPreProcessing is a function that wraps a dispatcher in a
// PreProcessingDispatcher
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
