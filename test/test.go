// handlers_test.go
package test

import (
	"goskel/http"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestHandleIndex(t *testing.T) {

	is := is.New(t)
	srv := http.NewServer()
	srvs, teardown := setup(t)
	defer teardown()
	srv.Service = srvs
	req, err := http.NewRequest("GET", "/", nil)
	is.NoErr(err)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusOK)
}

func TestNotFound(t *testing.T) {

	is := is.New(t)
	srv := http.NewServer()
	srvs, teardown := setup(t)
	defer teardown()
	srv.Service = srvs
	req, err := http.NewRequest("GET", "/notfound", nil)
	is.NoErr(err)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusNotFound)
}
