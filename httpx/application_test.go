package httpx_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"bitbucket.org/mparaiso/go-fizz/httpx"
)

func TestNewApplication(t *testing.T) {
	application := httpx.NewApplication()
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)
	application.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Index")
	}, httpx.WithMethod("GET"))
	application.ServeHTTP(response, request)
	Expect(t, response.Code, 200)
	Expect(t, response.Body.String(), "Index")
}

func Expect(t *testing.T, got, want interface{}) {
	t.Helper()
	if got != want {
		t.Fatalf("got: '%+v', want: '%+v'", got, want)
	}
}
