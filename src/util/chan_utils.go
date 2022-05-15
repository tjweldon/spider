package util

func Split[U any](input <-chan U, outputBufferLen int) (out1 <-chan U, out2 <-chan U) {
	worker := func(in <-chan U, outA, outB chan<- U) {
		defer close(outA)
		defer close(outB)

		for msg := range in {
			select {
			case outA <- msg:
				select {
				case outB <- msg:
				}
			case outB <- msg:
				select {
				case outA <- msg:
				}
			}
		}
	}

	r1, r2 := make(chan U, outputBufferLen), make(chan U, outputBufferLen)
	out1, out2 = r1, r2
	go worker(input, r1, r2)

	return out1, out2
}

func IsClosed[T any](channel <-chan T) bool {
	isClosed := false

	select {
	case _, open := <-channel:
		isClosed = !open
	default:
	}

	return isClosed
}

func AwaitClosure[T any](channel <-chan T) {
	if IsClosed[T](channel) {
		return
	}
	for range channel {
		// do nothing
	}
}

type Closer interface {
	Close()
}
