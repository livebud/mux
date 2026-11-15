package mux_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"

	"github.com/livebud/mux"
	"github.com/matryer/is"
	"github.com/matthewmueller/diff"
)

// Handler returns the raw query
func handler(route string) http.HandlerFunc {
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
	req := httptest.NewRequest(parts[0], parts[1], nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	actual, err := httputil.DumpResponse(rec.Result(), true)
	if err != nil {
		if err.Error() == expect {
			return
		}
		t.Fatal(err)
	}
	diff.TestHTTP(t, string(actual), expect)
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
	is := is.New(t)
	router := mux.New()
	router.Get("/", handler("GET /"))
	router.Get("/hi/", handler("GET /hi/"))
	err := router.Set("GET", "/hi", handler("GET /hi"))
	is.True(errors.Is(err, mux.ErrDuplicate))

	requestEqual(t, router, "GET /", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /
	`)
	requestEqual(t, router, "GET /hi", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /hi/
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
	is := is.New(t)
	router := mux.New()
	router.Get("/hi", handler("GET /hi"))
	// Routes must be lowercase; requests match case-insensitively
	is.True(errors.Is(router.Set("GET", "/HI", handler("GET /HI")), mux.ErrInvalid))
	is.True(errors.Is(router.Set("GET", "/Hi", handler("GET /Hi")), mux.ErrInvalid))
	is.True(errors.Is(router.Set("GET", "/hI", handler("GET /hI")), mux.ErrInvalid))
	is.True(errors.Is(router.Set("GET", "/HI/", handler("GET /HI/")), mux.ErrInvalid))
	is.True(errors.Is(router.Set("GET", "/hI/", handler("GET /hI/")), mux.ErrInvalid))
	is.True(errors.Is(router.Set("GET", "/Hi/", handler("GET /Hi/")), mux.ErrInvalid))
	is.True(errors.Is(router.Set("GET", "/hi/", handler("GET /hi/")), mux.ErrDuplicate))

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
		HTTP/1.1 405 Method Not Allowed
		Connection: close
		Allow: HEAD
		Content-Type: text/plain; charset=utf-8
		X-Content-Type-Options: nosniff

		Method Not Allowed
	`)
	requestEqual(t, router, "HEAD /10", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		HEAD /{id} id=10
	`)
}

func TestResource(t *testing.T) {
	router := mux.New()
	router.Set(http.MethodGet, "/{id}/edit", handler("GET /{id}/edit"))
	router.Set(http.MethodGet, "/", handler("GET /"))
	router.Set(http.MethodGet, "/new", handler("GET /new"))
	router.Set(http.MethodGet, "/{id}", handler("GET /{id}"))
	requestEqual(t, router, "GET /", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /
	`)
	requestEqual(t, router, "GET /10", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /{id} id=10
	`)
	requestEqual(t, router, "GET /new", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /new
	`)
	requestEqual(t, router, "GET /10/edit", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /{id}/edit id=10
	`)
}

func TestFind(t *testing.T) {
	is := is.New(t)
	router := mux.New()
	h := handler("GET /{id}")
	router.Get("/{id}", h)
	route, err := router.Find("GET", "/{id}")
	is.NoErr(err)
	is.Equal(route.Method, "GET")
	is.Equal(route.Route, "/{id}")
	is.Equal(route.Handler, h)
	route, err = router.Find("POST", "/{id}")
	is.True(errors.Is(err, mux.ErrNoMatch))
	is.Equal(route, nil)
	// Find resolves routes relative to the group, like Get does
	api := router.Group("/api")
	api.Get("/x", h)
	route, err = api.Find("GET", "/x")
	is.NoErr(err)
	is.Equal(route.Route, "/api/x")
}

func TestDuplicateError(t *testing.T) {
	is := is.New(t)
	router := mux.New()
	err := router.Set("GET", "/", handler("GET /"))
	is.NoErr(err)
	err = router.Set("GET", "/", handler("GET /"))
	is.True(errors.Is(err, mux.ErrDuplicate))
}

func TestDuplicatePanic(t *testing.T) {
	is := is.New(t)
	router := mux.New()
	router.Get("/", handler("GET /"))
	defer func() {
		err, ok := recover().(error)
		is.True(ok)
		is.True(errors.Is(err, mux.ErrDuplicate))
	}()
	router.Get("/", handler("GET /"))
}

func TestList(t *testing.T) {
	is := is.New(t)
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
	routes := router.Routes()
	is.Equal(len(routes), 17)
	is.Equal(routes[0].String(), "GET /")
	is.Equal(routes[1].String(), "GET /posts/{post_id}/comments")
	is.Equal(routes[2].String(), "GET /posts/{post_id}/comments/{id}.{format?}")
	is.Equal(routes[3].String(), "GET /posts/{post_id}/comments/{id}/edit")
	is.Equal(routes[4].String(), "GET /posts/{postid}/comments/new")
	is.Equal(routes[5].String(), "GET /users")
	is.Equal(routes[6].String(), "GET /users/new")
	is.Equal(routes[7].String(), "GET /users/{id}.{format?}")
	is.Equal(routes[8].String(), "GET /users/{id}/edit")
	is.Equal(routes[9].String(), "POST /posts/{post_id}/comments")
	is.Equal(routes[10].String(), "POST /users")
	is.Equal(routes[11].String(), "PUT /posts/{post_id}/comments/{id}.{format?}")
	is.Equal(routes[12].String(), "PUT /users/{id}.{format?}")
	is.Equal(routes[13].String(), "PATCH /posts/{post_id}/comments/{id}.{format?}")
	is.Equal(routes[14].String(), "PATCH /users/{id}.{format?}")
	is.Equal(routes[15].String(), "DELETE /posts/{post_id}/comments/{id}.{format?}")
	is.Equal(routes[16].String(), "DELETE /users/{id}.{format?}")
}

func TestListUnranked(t *testing.T) {
	is := is.New(t)
	router := mux.New()
	is.NoErr(router.Set(http.MethodOptions, "/p", handler("OPTIONS /p")))
	is.NoErr(router.Set(http.MethodHead, "/p", handler("HEAD /p")))
	router.Get("/p", handler("GET /p"))
	// Map iteration is random, so repeat to catch an unstable order
	for range 20 {
		routes := router.Routes()
		is.Equal(len(routes), 3)
		is.Equal(routes[0].String(), "GET /p")
		is.Equal(routes[1].String(), "HEAD /p")
		is.Equal(routes[2].String(), "OPTIONS /p")
	}
}

func TestMissingRoot(t *testing.T) {
	router := mux.New()
	router.Get("/signup", handler("GET /signup"))
	router.Post("/users", handler("POST /users"))
	router.Get("/login", handler("GET /login"))
	router.Post("/sessions", handler("POST /sessions"))
	router.Delete("/sessions", handler("DELETE /sessions"))
	router.Get("/", handler("GET /"))

	requestEqual(t, router, "GET /", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /
	`)
}

func TestMatch(t *testing.T) {
	is := is.New(t)
	router := mux.New()
	router.Get("/signup", handler("GET /signup"))
	router.Post("/users", handler("POST /users"))
	router.Get("/login", handler("GET /login"))
	router.Post("/sessions", handler("POST /sessions"))
	router.Delete("/sessions", handler("DELETE /sessions"))
	router.Get("/users/{user_id}/posts/{id}", handler("GET /users/{user_id}/posts/{id}"))
	router.Get("/", handler("GET /"))

	match, err := router.Match(http.MethodGet, "/")
	is.NoErr(err)
	is.Equal(match.Route, "/")
	is.Equal(match.Path, "/")
	is.Equal(len(match.Slots), 0)

	match, err = router.Match(http.MethodPost, "/")
	is.True(errors.Is(err, mux.ErrNoMatch))
	is.Equal(match, nil)

	match, err = router.Match(http.MethodGet, "/ok")
	is.True(errors.Is(err, mux.ErrNoMatch))
	is.Equal(match, nil)

	match, err = router.Match(http.MethodPost, "/ok")
	is.True(errors.Is(err, mux.ErrNoMatch))
	is.Equal(match, nil)

	match, err = router.Match(http.MethodPost, "/users")
	is.NoErr(err)
	is.Equal(match.Route, "/users")
	is.Equal(match.Path, "/users")
	is.Equal(len(match.Slots), 0)

	match, err = router.Match(http.MethodPost, "/users/")
	is.NoErr(err)
	is.Equal(match.Route, "/users")
	is.Equal(match.Path, "/users")
	is.Equal(len(match.Slots), 0)

	match, err = router.Match(http.MethodGet, "/users/10/posts/20")
	is.NoErr(err)
	is.Equal(match.Route, "/users/{user_id}/posts/{id}")
	is.Equal(match.Path, "/users/10/posts/20")
	is.Equal(len(match.Slots), 2)
	is.Equal(match.Slots[0].Key, "user_id")
	is.Equal(match.Slots[0].Value, "10")
	is.Equal(match.Slots[1].Key, "id")
	is.Equal(match.Slots[1].Value, "20")
}

func TestMiddleware(t *testing.T) {
	router := mux.New()
	router = router.Use(
		mux.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-A", "A")
				next.ServeHTTP(w, r)
				// Note: Can't use a header here because we've already written
				w.Write([]byte(" (a)"))
			})
		}),
		mux.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-B", "B")
				next.ServeHTTP(w, r)
				// Note: Can't use a header here because we've already written
				w.Write([]byte("(b)"))
			})
		}))
	router.Get("/", handler("GET /"))
	requestEqual(t, router, "GET /", `
			HTTP/1.1 200 OK
			Connection: close
			Content-Type: text/plain; charset=utf-8
			X-A: A
			X-B: B

			GET / (b) (a)
	`)
}

func TestRouteSpecificMiddleware(t *testing.T) {
	middleware := func(name string) mux.Middleware {
		return mux.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("(" + name + ") "))
				next.ServeHTTP(w, r)
				w.Write([]byte("(" + name + ") "))
			})
		})
	}

	router := mux.New()
	router = router.Use(middleware("a"), middleware("b"))
	router.Use(middleware("c")).Get("/", handler("GET /"))
	router.Post("/", handler("POST /"))

	requestEqual(t, router, "GET /", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		(a) (b) (c) GET / (c) (b) (a)
	`)
	requestEqual(t, router, "POST /", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		(a) (b) POST / (b) (a)
	`)
}

func TestComposeNone(t *testing.T) {
	router := mux.New()
	mw := mux.Compose()
	router = router.Use(mw)
	router.Get("/", handler("GET /"))
	requestEqual(t, router, "GET /", `
			HTTP/1.1 200 OK
			Connection: close
			Content-Type: text/plain; charset=utf-8

			GET /
	`)
}

func TestComposeOne(t *testing.T) {
	router := mux.New()
	mw := mux.Compose(
		mux.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-A", "A")
				next.ServeHTTP(w, r)
				// Note: Can't use a header here because we've already written
				w.Write([]byte("(a)"))
			})
		}),
	)
	router = router.Use(mw)
	router.Get("/", handler("GET /"))
	requestEqual(t, router, "GET /", `
			HTTP/1.1 200 OK
			Connection: close
			Content-Type: text/plain; charset=utf-8
			X-A: A

			GET / (a)
	`)
}
func TestComposeMany(t *testing.T) {
	router := mux.New()
	mw := mux.Compose(
		mux.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-A", "A")
				next.ServeHTTP(w, r)
				// Note: Can't use a header here because we've already written
				w.Write([]byte(" (a)"))
			})
		}),
		mux.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-B", "B")
				next.ServeHTTP(w, r)
				// Note: Can't use a header here because we've already written
				w.Write([]byte("(b)"))
			})
		}),
	)
	router = router.Use(mw)
	router.Get("/", handler("GET /"))
	requestEqual(t, router, "GET /", `
			HTTP/1.1 200 OK
			Connection: close
			Content-Type: text/plain; charset=utf-8
			X-A: A
			X-B: B

			GET / (b) (a)
	`)
}

func TestMiddlewareWrapsNonMatches(t *testing.T) {
	router := mux.New()
	router = router.Use(mux.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-A", "A")
			next.ServeHTTP(w, r)
		})
	}))
	router.Get("/", handler("GET /"))
	requestEqual(t, router, "GET /missing", `
			HTTP/1.1 404 Not Found
			Connection: close
			Content-Type: text/plain; charset=utf-8
			X-A: A
			X-Content-Type-Options: nosniff

			404 page not found
	`)
	requestEqual(t, router, "POST /", `
			HTTP/1.1 405 Method Not Allowed
			Connection: close
			Allow: GET, HEAD
			Content-Type: text/plain; charset=utf-8
			X-A: A
			X-Content-Type-Options: nosniff

			Method Not Allowed
	`)
}

func TestHeadFallsBackToGet(t *testing.T) {
	is := is.New(t)
	router := mux.New()
	router.Get("/users/{id}", handler("GET /users/{id}"))
	requestEqual(t, router, "HEAD /users/10", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /users/{id} id=10
	`)
	match, err := router.Match(http.MethodHead, "/users/10")
	is.NoErr(err)
	is.Equal(match.Method, http.MethodGet)
	// An explicit HEAD route wins over the GET fallback
	is.NoErr(router.Set(http.MethodHead, "/users/{id}", handler("HEAD /users/{id}")))
	requestEqual(t, router, "HEAD /users/10", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		HEAD /users/{id} id=10
	`)
}

func TestMethodNotAllowed(t *testing.T) {
	router := mux.New()
	router.Get("/users/{id}", handler("GET /users/{id}"))
	router.Delete("/users/{id}", handler("DELETE /users/{id}"))
	router.Post("/users", handler("POST /users"))
	requestEqual(t, router, "PUT /users/10", `
		HTTP/1.1 405 Method Not Allowed
		Connection: close
		Allow: DELETE, GET, HEAD
		Content-Type: text/plain; charset=utf-8
		X-Content-Type-Options: nosniff

		Method Not Allowed
	`)
}

func TestMiddlewareWrongMethodCallsNext(t *testing.T) {
	router := mux.New()
	router.Get("/", handler("GET /"))
	next := router.Middleware(handler("next"))
	requestEqual(t, next, "POST /", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		next 
	`)
}

func TestPostBody(t *testing.T) {
	is := is.New(t)
	router := mux.New()
	router.Post("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		is.NoErr(err)
		w.Write(body)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"jon"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	is.Equal(rec.Code, http.StatusOK)
	is.Equal(rec.Body.String(), `{"name":"jon"}`)
}

func TestWildcardFallback(t *testing.T) {
	router := mux.New()
	router.Get("/", handler("GET /"))
	router.Get("/{public?}", handler("GET /{public?}"))
	router.Get("/{public*}", handler("GET /{public*}"))
	requestEqual(t, router, "GET /", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /
	`)
	requestEqual(t, router, "GET /index.css", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /{public?} public=index.css
	`)
	requestEqual(t, router, "GET /assets/index.css", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /{public*} public=assets%2Findex.css
	`)
}

func TestWellKnown(t *testing.T) {
	router := mux.New()
	router.Get("/", handler("GET /"))
	requestEqual(t, router, "GET /.well-known/appspecific/com.chrome.devtools.json", `
		HTTP/1.1 404 Not Found
		Connection: close
		Content-Type: text/plain; charset=utf-8
		X-Content-Type-Options: nosniff

		404 page not found
	`)
	router.Get("/.well-known/{path*}", handler("GET /.well-known/{path*}"))
	requestEqual(t, router, "GET /.well-known/openid-configuration", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /.well-known/{path*} path=openid-configuration
	`)
}

func TestGroup(t *testing.T) {
	router := mux.New()
	auth := router.Group("/auth")
	auth.Get("/login", handler("GET /auth/login"))
	auth.Post("/login", handler("POST /auth/login"))
	auth.Get("/logout", handler("GET /auth/logout"))

	v1 := router.Group("/api/{version}")
	v1.Get("/users", handler("GET /api/{version}/users"))
	v1.Post("/users", handler("POST /api/{version}/users"))
	v1.Get("/users/{id}", handler("GET /api/{version}/users/{id}"))

	requestEqual(t, router, "GET /auth/login", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /auth/login
	`)
	requestEqual(t, router, "POST /auth/login", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		POST /auth/login
	`)
	requestEqual(t, router, "GET /auth/logout", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /auth/logout
	`)
	requestEqual(t, router, "GET /api/v1/users", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /api/{version}/users version=v1
	`)
	requestEqual(t, router, "POST /api/v1/users", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		POST /api/{version}/users version=v1
	`)
	requestEqual(t, router, "GET /api/v1/users/10", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /api/{version}/users/{id} id=10&version=v1
	`)
}

type slackHandler struct {
}

var _ mux.Mountable = (*slackHandler)(nil)

func (s *slackHandler) Mount(router *mux.Router) {
	router.Post("/events", handler("POST /slack/events"))
	router.Get("/commands", handler("GET /slack/commands"))
}

type telegramHandler struct {
}

var _ mux.Mountable = (*telegramHandler)(nil)

func (s *telegramHandler) Mount(router *mux.Router) {
	router.Post("/events", handler("POST /telegram/events"))
	router.Get("/commands", handler("GET /telegram/commands"))
}

func TestMount(t *testing.T) {
	router := mux.New()
	router.Mount(&slackHandler{})

	requestEqual(t, router, "POST /events", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		POST /slack/events
	`)
	requestEqual(t, router, "GET /commands", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /slack/commands
	`)
}

func TestGroupMount(t *testing.T) {
	router := mux.New()
	router.Group("/slack").Mount(&slackHandler{})
	router.Group("/telegram").Mount(&telegramHandler{})

	requestEqual(t, router, "POST /slack/events", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		POST /slack/events
	`)
	requestEqual(t, router, "GET /slack/commands", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /slack/commands
	`)
	requestEqual(t, router, "POST /telegram/events", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		POST /telegram/events
	`)
	requestEqual(t, router, "GET /telegram/commands", `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/plain; charset=utf-8

		GET /telegram/commands
	`)
}
