package logger_test

import (
	"fmt"
	"testing"

	"github.com/itsubaki/logger"
)

func TestMustPanic(t *testing.T) {
	defer func() {
		if rec := recover(); rec != nil {
			err, ok := rec.(error)
			if !ok {
				t.Fail()
			}

			if err.Error() != "something went wrong" {
				t.Fail()
			}
		}
	}()

	logger.Must(nil, fmt.Errorf("something went wrong"))
	t.Fail()
}
