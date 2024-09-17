package misc

import (
	"log"
	"net/http"
)

// HumanReadableError represents error information
// that can be fed back to a human user.
//
// This prevents internal state that might be sensitive
// being leaked to the outside world.
type HumanReadableError interface {
	error
	HumanError() string
	HTTPCode() int
}
type HumanReadableWrapper struct {
	error
	ToHuman string
	Code    int
}

func (h HumanReadableWrapper) HumanError() string { return h.ToHuman }
func (h HumanReadableWrapper) HTTPCode() int      { return h.Code }
func NewHumanReadableError(toHuman string, code int, err error) HumanReadableError {
	return HumanReadableWrapper{
		error:   err,
		ToHuman: toHuman,
		Code:    code,
	}
}

type HandlerWithErr func(http.ResponseWriter, *http.Request) error

func AnnotateError(h HandlerWithErr) HandlerWithErr {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		// parse POST body, limit request size
		if err = r.ParseForm(); err != nil {
			return HumanReadableWrapper{
				ToHuman: "Something went wrong! Please try again.",
				Code:    http.StatusBadRequest,
				error:   err,
			}
		}

		return h(w, r)
	}
}

func ErrorHandling(handler HandlerWithErr) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			var errorString string = "Something went wrong! Please try again."
			var errorCode int = 500

			if v, ok := err.(HumanReadableError); ok {
				errorString, errorCode = v.HumanError(), v.HTTPCode()
			}

			log.Println(err)
			w.WriteHeader(errorCode)
			w.Write([]byte(errorString))
			return
		}
	})
}

func Middleware(h HandlerWithErr) http.Handler {
	return ErrorHandling(AnnotateError(h))
}
