package router_test

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"

	router "github.com/livebud/mux"
	"github.com/matthewmueller/diff"
)

type test struct {
	routes   []*route
	requests []*request
}

type route struct {
	method string
	route  string
	err    string
}

type request struct {
	method string
	path   string

	// response
	status   int
	location string
	body     string
}

// Handler returns the raw query
func handler(route string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.URL.RawQuery))
	})
}

func requestEqual(t testing.TB, router http.Handler, request string, expect string) {
	t.Helper()
	parts := strings.SplitN(request, " ", 2)
	if len(parts) != 2 {
		t.Fatalf("invalid request: %s", request)
	}
	u, err := url.Parse(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(parts[0], u.Path, nil)
	req.URL.RawQuery = u.RawQuery
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	actual, err := httputil.DumpResponse(rec.Result(), true)
	if err != nil {
		if err.Error() == expect {
			return
		}
		t.Fatal(err)
	}
	diff.TestHTTP(t, expect, string(actual))
}

func TestSanity(t *testing.T) {
	router := router.New()
	router.Get("/", handler("/"))
	requestEqual(t, router, "GET /", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8
	`)
	requestEqual(t, router, "GET /hi", `
		HTTP/1.1 404 Not Found
		Connection: close
		Content-Type: text/plain; charset=utf-8
		X-Content-Type-Options: nosniff

		404 page not found
	`)
	router.Get("/{name}", handler("/{name}"))
	requestEqual(t, router, "GET /anki", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		name=anki
	`)
}

func TestREST(t *testing.T) {
	router := router.New()
	router.Get("/", handler("/"))
	router.Get("/users", handler("/users"))
	router.Get("/users/new", handler("/users/new"))
	router.Post("/users", handler("/users"))
	router.Get("/users/{id}.{format?}", handler("/users/{id}.{format?}"))
	router.Get("/users/{id}/edit", handler("/users/{id}/edit"))
	router.Patch("/users/{id}.{format?}", handler("/users/{id}.{format?}"))
	router.Put("/users/{id}.{format?}", handler("/users/{id}.{format?}"))
	router.Delete("/users/{id}.{format?}", handler("/users/{id}.{format?}"))
	router.Get("/posts/{post_id}/comments", handler("/posts/{post_id}/comments"))
	router.Get("/posts/{postid}/comments/new", handler("/posts/{post_id}/comments/new"))
	router.Post("/posts/{post_id}/comments", handler("/posts/{post_id}/comments"))
	router.Get("/posts/{post_id}/comments/{id}.{format?}", handler("/posts/{post_id}/comments/{id}.{format?}"))
	router.Get("/posts/{post_id}/comments/{id}/edit", handler("/posts/{post_id}/comments/{id}/edit"))
	router.Patch("/posts/{post_id}/comments/{id}.{format?}", handler("/posts/{post_id}/comments/{id}.{format?}"))
	router.Put("/posts/{post_id}/comments/{id}.{format?}", handler("/posts/{post_id}/comments/{id}.{format?}"))
	router.Delete("/posts/{post_id}/comments/{id}.{format?}", handler("/posts/{post_id}/comments/{id}.{format?}"))

	// requests
	requestEqual(t, router, "GET /", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8
	`)
	requestEqual(t, router, "GET /users", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8
	`)
	requestEqual(t, router, "GET /users/new", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8
	`)
	requestEqual(t, router, "POST /users", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8
	`)
	requestEqual(t, router, "GET /users/10", `
		HTTP/1.1 404 Not Found
		Connection: close
		Content-Type: text/plain; charset=utf-8
		X-Content-Type-Options: nosniff

		404 page not found
	`)
	requestEqual(t, router, "GET /users/10.json", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=json&id=10
	`)
	requestEqual(t, router, "GET /users/10.rss", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=rss&id=10
	`)
	requestEqual(t, router, "GET /users/10.html", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=html&id=10
	`)
	requestEqual(t, router, "GET /users/10/edit", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		id=10
	`)
	requestEqual(t, router, "PATCH /users/10", `
		HTTP/1.1 404 Not Found
		Connection: close
		Content-Type: text/plain; charset=utf-8
		X-Content-Type-Options: nosniff

		404 page not found
	`)
	requestEqual(t, router, "PATCH /users/10.json", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=json&id=10
	`)
	requestEqual(t, router, "PATCH /users/10.rss", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=rss&id=10
	`)
	requestEqual(t, router, "PATCH /users/10.html", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=html&id=10
	`)
	requestEqual(t, router, "PUT /users/10", `
		HTTP/1.1 404 Not Found
		Connection: close
		Content-Type: text/plain; charset=utf-8
		X-Content-Type-Options: nosniff

		404 page not found
	`)
	requestEqual(t, router, "PUT /users/10.json", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=json&id=10
	`)
	requestEqual(t, router, "PUT /users/10.rss", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=rss&id=10
	`)
	requestEqual(t, router, "PUT /users/10.html", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=html&id=10
	`)
	requestEqual(t, router, "DELETE /users/10", `
		HTTP/1.1 404 Not Found
		Connection: close
		Content-Type: text/plain; charset=utf-8
		X-Content-Type-Options: nosniff

		404 page not found
	`)
	requestEqual(t, router, "DELETE /users/10.json", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=json&id=10
	`)
	requestEqual(t, router, "DELETE /users/10.rss", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=rss&id=10
	`)
	requestEqual(t, router, "DELETE /users/10.html", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=html&id=10
	`)
	requestEqual(t, router, "GET /posts/1/comments", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		post_id=1
	`)
	requestEqual(t, router, "GET /posts/1/comments/new", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		postid=1
	`)
	requestEqual(t, router, "POST /posts/1/comments", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		post_id=1
	`)
	requestEqual(t, router, "GET /posts/1/comments/2", `
		HTTP/1.1 404 Not Found
		Connection: close
		Content-Type: text/plain; charset=utf-8
		X-Content-Type-Options: nosniff

		404 page not found
	`)
	requestEqual(t, router, "GET /posts/1/comments/2.json", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=json&id=2&post_id=1
	`)
	requestEqual(t, router, "GET /posts/1/comments/2.rss", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=rss&id=2&post_id=1
	`)
	requestEqual(t, router, "GET /posts/1/comments/2.html", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=html&id=2&post_id=1
	`)
	requestEqual(t, router, "GET /posts/1/comments/2/edit", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		id=2&post_id=1
	`)
	requestEqual(t, router, "PATCH /posts/1/comments/2", `
		HTTP/1.1 404 Not Found
		Connection: close
		Content-Type: text/plain; charset=utf-8
		X-Content-Type-Options: nosniff

		404 page not found
	`)
	requestEqual(t, router, "PATCH /posts/1/comments/2.json", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=json&id=2&post_id=1
	`)
	requestEqual(t, router, "PATCH /posts/1/comments/2.rss", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=rss&id=2&post_id=1
	`)
	requestEqual(t, router, "PATCH /posts/1/comments/2.html", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=html&id=2&post_id=1
	`)
	requestEqual(t, router, "PUT /posts/1/comments/2", `
		HTTP/1.1 404 Not Found
		Connection: close
		Content-Type: text/plain; charset=utf-8
		X-Content-Type-Options: nosniff

		404 page not found
	`)
	requestEqual(t, router, "PUT /posts/1/comments/2.json", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=json&id=2&post_id=1
	`)
	requestEqual(t, router, "PUT /posts/1/comments/2.rss", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=rss&id=2&post_id=1
	`)
	requestEqual(t, router, "PUT /posts/1/comments/2.html", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=html&id=2&post_id=1
	`)
	requestEqual(t, router, "DELETE /posts/1/comments/2", `
		HTTP/1.1 404 Not Found
		Connection: close
		Content-Type: text/plain; charset=utf-8
		X-Content-Type-Options: nosniff

		404 page not found
	`)
	requestEqual(t, router, "DELETE /posts/1/comments/2.json", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=json&id=2&post_id=1
	`)
	requestEqual(t, router, "DELETE /posts/1/comments/2.rss", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=rss&id=2&post_id=1
	`)
	requestEqual(t, router, "DELETE /posts/1/comments/2.html", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=html&id=2&post_id=1
	`)
}

func TestSlotPriority(t *testing.T) {
	router := router.New()
	router.Get("/", handler("/"))
	router.Get("/users/{id}", handler("/users/{id}"))
	router.Get("/users/{id}.{format?}", handler("/users/{id}.{format?}"))
	router.Get("/posts/{post_id}/comments/{id}.{format?}", handler("/posts/{post_id}/comments/{id}.{format?}"))

	requestEqual(t, router, "GET /?id=10", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		id=10
	`)
	requestEqual(t, router, `GET /users/10?id=20&format=bin&other=true`, `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=bin&id=10&other=true
	`)
	requestEqual(t, router, `GET /users/10.json?id=20&format=bin&other=true`, `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=json&id=10&other=true
	`)
	requestEqual(t, router, `GET /posts/10/comments/20.json?id=30&post_id=30&format=bin&other=true`, `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		format=json&id=20&other=true&post_id=10
	`)
}

func TestTrailingSlash(t *testing.T) {
	router := router.New()
	router.Get("/", handler("/"))
	router.Get("/hi/", handler("/hi/"))
	router.Get("/hi", handler("/hi"))

	requestEqual(t, router, "GET /", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8
	`)
	requestEqual(t, router, "GET /hi/", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8
	`)
	requestEqual(t, router, "GET /hi///", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8
	`)
}

func TestInsensitive(t *testing.T) {
	router := router.New()
	router.Get("/HI", handler("/HI"))
	router.Get("/hi", handler("/hi"))
	router.Get("/Hi", handler("/Hi"))
	router.Get("/hI", handler("/hI"))
	router.Get("/HI/", handler("/HI/"))
	router.Get("/hi/", handler("/hi/"))
	router.Get("/hI/", handler("/hI/"))
	router.Get("/Hi/", handler("/Hi/"))

	requestEqual(t, router, "GET /hi", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8
	`)
	requestEqual(t, router, "GET /HI", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8
	`)
	requestEqual(t, router, "GET /Hi", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8
	`)
	requestEqual(t, router, "GET /hi/", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8
	`)
	requestEqual(t, router, "GET /HI/", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8
	`)
	requestEqual(t, router, "GET /Hi/", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8
	`)
	requestEqual(t, router, "GET /hI/", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8
	`)
	requestEqual(t, router, "GET /HI////", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8
	`)
}

func TestSet(t *testing.T) {
	router := router.New()
	router.Set(http.MethodHead, "/{id}", handler("/{id}"))
	requestEqual(t, router, "GET /10", `
		HTTP/1.1 404 Not Found
		Connection: close
		Content-Type: text/plain; charset=utf-8
		X-Content-Type-Options: nosniff

		404 page not found
	`)
	requestEqual(t, router, "HEAD /10", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		id=10
	`)
}
