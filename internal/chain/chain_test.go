package chain_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/livebud/mux/internal/chain"
	"github.com/matryer/is"
)

func TestChainOne(t *testing.T) {
	is := is.New(t)
	handler := chain.All(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("a", "aa")
			body, err := io.ReadAll(r.Body)
			is.NoErr(err)
			w.Write([]byte("<a>" + string(body) + "</a>"))
		}),
	)
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	res := w.Result()
	is.Equal(res.StatusCode, http.StatusOK)
	is.Equal(res.Header.Get("a"), "aa")
	body, err := io.ReadAll(res.Body)
	is.NoErr(err)
	is.Equal(string(body), "<a></a>")
}

func TestChain(t *testing.T) {
	is := is.New(t)
	handler := chain.All(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("a", "aa")
			body, err := io.ReadAll(r.Body)
			is.NoErr(err)
			w.Write([]byte("<a>" + string(body) + "</a>"))
		}),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("b", "bb")
			body, err := io.ReadAll(r.Body)
			is.NoErr(err)
			w.Write([]byte("<b>" + string(body) + "</b>"))
		}),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("b", "cc")
			body, err := io.ReadAll(r.Body)
			is.NoErr(err)
			w.Write([]byte("<c>" + string(body) + "</c>"))
		}),
	)
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	res := w.Result()
	is.Equal(res.StatusCode, http.StatusOK)
	is.Equal(res.Header.Get("a"), "aa")
	is.Equal(res.Header.Get("b"), "cc")
	body, err := io.ReadAll(res.Body)
	is.NoErr(err)
	is.Equal(string(body), "<c><b><a></a></b></c>")
}
