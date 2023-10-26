package tofu

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type BetterError struct {
	error
	code int
}

func ApiHandleError(h func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}
		HandleError(w, r, err)
	}
}
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	betterError := err.(BetterError)
	if betterError.code == 0 {
		betterError.code = http.StatusOK
	}
	fmt.Println("[[ERORR]] - ", betterError.error)
	handleStats(r, betterError.code)
	writeJSON(w, r, betterError.code, betterError.error)
}

func HandleSuccess(w http.ResponseWriter, r *http.Request, statusCode int, body interface{}) {
	handleStats(r, statusCode)
	writeJSON(w, r, statusCode, body)
}

func handleStats(r *http.Request, statusCode int) {
	if 0 == statusCode {
		return
	}
}

func getValidPath(r *http.Request) string {
	rawPatterns := chi.RouteContext(r.Context()).RoutePatterns
	lastPattern := rawPatterns[len(rawPatterns)-1]

	lastPattern = strings.TrimPrefix(lastPattern, "/")
	lastPattern = strings.Replace(lastPattern, "/", "_", -1)

	return dropSubstrings([]string{"{", "}"}, lastPattern)
}
func dropSubstrings(substrings []string, haystack string) string {
	for _, substring := range substrings {
		haystack = strings.Replace(haystack, substring, "", -1)
	}

	return haystack
}

func writeJSON(w http.ResponseWriter, r *http.Request, statusCode int, body interface{}) {

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if body != nil {
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(true)
		if err := enc.Encode(body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func Wrap(msg string, err error) error {
	betterError := err.(BetterError)
	return BetterError{
		code:  betterError.code,
		error: errors.Join(fmt.Errorf(msg), betterError.error),
	}
}

func NewForbidden(msg string) error {
	return BetterError{
		code:  403,
		error: fmt.Errorf(msg),
	}
}

func NewBadRequest(msg string) error {
	return BetterError{
		code:  400,
		error: fmt.Errorf(msg),
	}
}

func NewInternalf(msg string, err error) error {
	return BetterError{
		code:  500,
		error: fmt.Errorf("%s -> %s", msg, err),
	}
}

func NewInternal(msg string) error {
	return BetterError{
		code:  500,
		error: fmt.Errorf(msg),
	}
}
