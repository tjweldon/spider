package util

type ErrEmpty struct{}

func (e ErrEmpty) Error() string {
	return "The deque is empty"
}

type Deque[T any] struct {
	members []T
}

func NewDeque[T any]() *Deque[T] {
	return &Deque[T]{[]T{}}
}

func (d *Deque[T]) Insert(item T) {
	d.members = append(
		d.members,
		item,
	)
}

func (d *Deque[T]) IsEmpty() bool {
	return len(d.members) == 0
}

func (d *Deque[T]) TakeOne() T {
	item := d.members[0]
	if len(d.members) > 1 {
		d.members = append([]T{}, d.members[1:]...)
	} else {
		d.members = []T{}
	}
	return item
}
