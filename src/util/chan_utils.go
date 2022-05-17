package util

// Split is a function that takes in the read side of a channel and duplexes it.
// With a zero size output buffer, the consumers are forced to consume at the
// same rate. The bigger the buffer is, the more one consumer can outpace the
// other, however it is ultimately always limited by the buffer size.
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

// IsClosed is a convenience function that takes the read side of a channel
// and returns true if it's closed.
func IsClosed[T any](channel <-chan T) bool {
	isClosed := false

	select {
	case _, open := <-channel:
		isClosed = !open
	default:
	}

	return isClosed
}

// AwaitClosure is a function that blocks while it consumes and discards
// messages from a channel until it is closed, at which point the function
// returns.
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
