package mux_test

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"

	"github.com/livebud/mux"
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
		w.Write([]byte(route + " " + r.URL.RawQuery))
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
	router := mux.New()
	router.Get("/", handler("GET /"))
	requestEqual(t, router, "GET /", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /
	`)
	requestEqual(t, router, "GET /hi", `
		HTTP/1.1 404 Not Found
		Connection: close
		Content-Type: text/plain; charset=utf-8
		X-Content-Type-Options: nosniff

		404 page not found
	`)
	router.Get("/{name}", handler("GET /{name}"))
	requestEqual(t, router, "GET /anki", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /{name} name=anki
	`)
}

func TestSample(t *testing.T) {
	router := mux.New()
	router.Get("/", handler("GET /"))
	router.Get("/users/{id}", handler("GET /users/{id}"))
	router.Post("/users/{id}.{format}", handler("POST /users/{id}.{format}"))
	router.Get("/posts/{post_id}/comments/{id}", handler("GET /posts/{post_id}/comments/{id}"))
	router.Get("/fly/{from}-{to}", handler("GET /fly/{from}-{to}"))
	router.Get("/v{major|[0-9]+}.{minor|[0-9]+}", handler("GET /v{major|[0-9]+}.{minor|[0-9]+}"))
	router.Get("/{owner}/{repo}/{branch}/{path*}", handler("GET /{owner}/{repo}/{branch}/{path*}"))
	requestEqual(t, router, "GET /", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /
	`)
	requestEqual(t, router, "GET /users/1", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /users/{id} id=1
	`)
	requestEqual(t, router, "POST /users/1.json", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		POST /users/{id}.{format} format=json&id=1
	`)
	requestEqual(t, router, "GET /posts/1/comments/1", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /posts/{post_id}/comments/{id} id=1&post_id=1
	`)
	requestEqual(t, router, "GET /fly/sfo-lax", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /fly/{from}-{to} from=sfo&to=lax
	`)
	requestEqual(t, router, "GET /v1.0", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /v{major|[0-9]+}.{minor|[0-9]+} major=1&minor=0
	`)
	requestEqual(t, router, "GET /1.0", `
		HTTP/1.1 404 Not Found
		Connection: close
		Content-Type: text/plain; charset=utf-8
		X-Content-Type-Options: nosniff

		404 page not found
	`)
	requestEqual(t, router, "GET /v1.a", `
		HTTP/1.1 404 Not Found
		Connection: close
		Content-Type: text/plain; charset=utf-8
		X-Content-Type-Options: nosniff

		404 page not found
	`)
	requestEqual(t, router, "GET /livebud/mux/main/path/to/file.go", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /{owner}/{repo}/{branch}/{path*} branch=main&owner=livebud&path=path%2Fto%2Ffile.go&repo=mux
	`)
}

func TestREST(t *testing.T) {
	router := mux.New()
	router.Get("/", handler("GET /"))
	router.Get("/users", handler("GET /users"))
	router.Get("/users/new", handler("GET /users/new"))
	router.Post("/users", handler("POST /users"))
	router.Get("/users/{id}.{format?}", handler("GET /users/{id}.{format?}"))
	router.Get("/users/{id}/edit", handler("GET /users/{id}/edit"))
	router.Patch("/users/{id}.{format?}", handler("PATCH /users/{id}.{format?}"))
	router.Put("/users/{id}.{format?}", handler("PUT /users/{id}.{format?}"))
	router.Delete("/users/{id}.{format?}", handler("DELETE /users/{id}.{format?}"))
	router.Get("/posts/{post_id}/comments", handler("GET /posts/{post_id}/comments"))
	router.Get("/posts/{postid}/comments/new", handler("GET /posts/{postid}/comments/new"))
	router.Post("/posts/{post_id}/comments", handler("POST /posts/{post_id}/comments"))
	router.Get("/posts/{post_id}/comments/{id}.{format?}", handler("GET /posts/{post_id}/comments/{id}.{format?}"))
	router.Get("/posts/{post_id}/comments/{id}/edit", handler("GET /posts/{post_id}/comments/{id}/edit"))
	router.Patch("/posts/{post_id}/comments/{id}.{format?}", handler("PATCH /posts/{post_id}/comments/{id}.{format?}"))
	router.Put("/posts/{post_id}/comments/{id}.{format?}", handler("PUT /posts/{post_id}/comments/{id}.{format?}"))
	router.Delete("/posts/{post_id}/comments/{id}.{format?}", handler("DELETE /posts/{post_id}/comments/{id}.{format?}"))

	// requests
	requestEqual(t, router, "GET /", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /
	`)
	requestEqual(t, router, "GET /users", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /users
	`)
	requestEqual(t, router, "GET /users/new", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /users/new
	`)
	requestEqual(t, router, "POST /users", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		POST /users
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

		GET /users/{id}.{format?} format=json&id=10
	`)
	requestEqual(t, router, "GET /users/10.rss", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /users/{id}.{format?} format=rss&id=10
	`)
	requestEqual(t, router, "GET /users/10.html", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /users/{id}.{format?} format=html&id=10
	`)
	requestEqual(t, router, "GET /users/10/edit", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /users/{id}/edit id=10
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

		PATCH /users/{id}.{format?} format=json&id=10
	`)
	requestEqual(t, router, "PATCH /users/10.rss", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		PATCH /users/{id}.{format?} format=rss&id=10
	`)
	requestEqual(t, router, "PATCH /users/10.html", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		PATCH /users/{id}.{format?} format=html&id=10
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

		PUT /users/{id}.{format?} format=json&id=10
	`)
	requestEqual(t, router, "PUT /users/10.rss", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		PUT /users/{id}.{format?} format=rss&id=10
	`)
	requestEqual(t, router, "PUT /users/10.html", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		PUT /users/{id}.{format?} format=html&id=10
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

		DELETE /users/{id}.{format?} format=json&id=10
	`)
	requestEqual(t, router, "DELETE /users/10.rss", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		DELETE /users/{id}.{format?} format=rss&id=10
	`)
	requestEqual(t, router, "DELETE /users/10.html", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		DELETE /users/{id}.{format?} format=html&id=10
	`)
	requestEqual(t, router, "GET /posts/1/comments", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /posts/{post_id}/comments post_id=1
	`)
	requestEqual(t, router, "GET /posts/1/comments/new", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /posts/{postid}/comments/new postid=1
	`)
	requestEqual(t, router, "POST /posts/1/comments", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		POST /posts/{post_id}/comments post_id=1
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

		GET /posts/{post_id}/comments/{id}.{format?} format=json&id=2&post_id=1
	`)
	requestEqual(t, router, "GET /posts/1/comments/2.rss", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /posts/{post_id}/comments/{id}.{format?} format=rss&id=2&post_id=1
	`)
	requestEqual(t, router, "GET /posts/1/comments/2.html", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /posts/{post_id}/comments/{id}.{format?} format=html&id=2&post_id=1
	`)
	requestEqual(t, router, "GET /posts/1/comments/2/edit", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /posts/{post_id}/comments/{id}/edit id=2&post_id=1
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

		PATCH /posts/{post_id}/comments/{id}.{format?} format=json&id=2&post_id=1
	`)
	requestEqual(t, router, "PATCH /posts/1/comments/2.rss", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		PATCH /posts/{post_id}/comments/{id}.{format?} format=rss&id=2&post_id=1
	`)
	requestEqual(t, router, "PATCH /posts/1/comments/2.html", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		PATCH /posts/{post_id}/comments/{id}.{format?} format=html&id=2&post_id=1
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

		PUT /posts/{post_id}/comments/{id}.{format?} format=json&id=2&post_id=1
	`)
	requestEqual(t, router, "PUT /posts/1/comments/2.rss", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		PUT /posts/{post_id}/comments/{id}.{format?} format=rss&id=2&post_id=1
	`)
	requestEqual(t, router, "PUT /posts/1/comments/2.html", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		PUT /posts/{post_id}/comments/{id}.{format?} format=html&id=2&post_id=1
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

		DELETE /posts/{post_id}/comments/{id}.{format?} format=json&id=2&post_id=1
	`)
	requestEqual(t, router, "DELETE /posts/1/comments/2.rss", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		DELETE /posts/{post_id}/comments/{id}.{format?} format=rss&id=2&post_id=1
	`)
	requestEqual(t, router, "DELETE /posts/1/comments/2.html", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		DELETE /posts/{post_id}/comments/{id}.{format?} format=html&id=2&post_id=1
	`)
}

func TestSlotPriority(t *testing.T) {
	router := mux.New()
	router.Get("/", handler("GET /"))
	router.Get("/users/{id}", handler("GET /users/{id}"))
	router.Get("/users/{id}.{format?}", handler("GET /users/{id}.{format?}"))
	router.Get("/posts/{post_id}/comments/{id}.{format?}", handler("GET /posts/{post_id}/comments/{id}.{format?}"))

	requestEqual(t, router, "GET /?id=10", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET / id=10
	`)
	requestEqual(t, router, `GET /users/10?id=20&format=bin&other=true`, `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /users/{id} format=bin&id=10&other=true
	`)
	requestEqual(t, router, `GET /users/10.json?id=20&format=bin&other=true`, `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /users/{id}.{format?} format=json&id=10&other=true
	`)
	requestEqual(t, router, `GET /posts/10/comments/20.json?id=30&post_id=30&format=bin&other=true`, `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /posts/{post_id}/comments/{id}.{format?} format=json&id=20&other=true&post_id=10
	`)
}

func TestTrailingSlash(t *testing.T) {
	router := mux.New()
	router.Get("/", handler("GET /"))
	router.Get("/hi/", handler("GET /hi/"))
	router.Get("/hi", handler("GET /hi"))

	requestEqual(t, router, "GET /", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /
	`)
	requestEqual(t, router, "GET /hi/", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /hi/
	`)
	requestEqual(t, router, "GET /hi///", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /hi/
	`)
}

func TestInsensitive(t *testing.T) {
	router := mux.New()
	router.Get("/HI", handler("GET /HI"))
	router.Get("/hi", handler("GET /hi"))
	router.Get("/Hi", handler("GET /Hi"))
	router.Get("/hI", handler("GET /hI"))
	router.Get("/HI/", handler("GET /HI/"))
	router.Get("/hi/", handler("GET /hi/"))
	router.Get("/hI/", handler("GET /hI/"))
	router.Get("/Hi/", handler("GET /Hi/"))

	requestEqual(t, router, "GET /hi", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /hi
	`)
	requestEqual(t, router, "GET /HI", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /hi
	`)
	requestEqual(t, router, "GET /Hi", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /hi
	`)
	requestEqual(t, router, "GET /hi/", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /hi
	`)
	requestEqual(t, router, "GET /HI/", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /hi
	`)
	requestEqual(t, router, "GET /Hi/", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /hi
	`)
	requestEqual(t, router, "GET /hI/", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /hi
	`)
	requestEqual(t, router, "GET /HI////", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /hi
	`)
}

func TestSet(t *testing.T) {
	router := mux.New()
	router.Set(http.MethodHead, "/{id}", handler("HEAD /{id}"))
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

		HEAD /{id} id=10
	`)
}
