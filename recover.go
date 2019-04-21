package pogo

import "github.com/pkg/errors"

func recoverHandledError(fn func()) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			type withStacktrace interface {
				error
				StackTrace() errors.StackTrace
			}
			if e, ok := rec.(withStacktrace); ok {
				err = e
			} else {
				panic(rec)
			}
		}
	}()

	fn()

	return nil
}
