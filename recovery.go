package servicefoundation

import (
	"fmt"
	"net/http"

	"github.com/Travix-International/logger"

	"runtime"
)

type recovery struct {
	loggy     *logger.Logger
	stackSize int
}

func newRecovery(loggy *logger.Logger) *recovery {
	return &recovery{
		loggy:     loggy,
		stackSize: 1024 * 8,
	}
}

func (rec *recovery) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer func() {
		if err := recover(); err != nil {
			if rw.Header().Get("Content-Type") == "" {
				rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
			}

			rw.WriteHeader(http.StatusInternalServerError)

			stack := make([]byte, rec.stackSize)
			stack = stack[:runtime.Stack(stack, false)]

			f := "PANIC: %s\n%s"
			rec.loggy.Error("RecoveryServerHTTPError", fmt.Sprintf(f, err, stack))
		}
	}()

	next(rw, r)
}
