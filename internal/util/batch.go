package util

import "errors"

var Batch batchutil

type batchutil struct{}

// BatchFunc is called for each batch.
// Any error will cancel the batching operation but returning Abort
// indicates it was deliberate, and not an error case.
type BatchFunc func(start, end int) error

// ErrAbort is a sentinal error as defined by Dave Cheney (@davecheney) which
// indicates a batch operation should abort early.
var ErrAbort = errors.New("done")

// All calls eachFn for all items
// Returns any error from eachFn except for Abort it returns nil.
func (b batchutil) All(count, batchSize int, eachFn BatchFunc) error {
	i := 0
	for i < count {
		end := i + batchSize - 1
		if end > count-1 {
			// passed end, so set to end item
			end = count - 1
		}
		err := eachFn(i, end)
		if err == ErrAbort {
			return nil
		}
		if err != nil {
			return err
		}
		i = end + 1
	}
	return nil
}
