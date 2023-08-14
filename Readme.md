# Mux

[![Go Reference](https://pkg.go.dev/badge/github.com/livebud/mux.svg)](https://pkg.go.dev/github.com/livebud/mux)

A minimal but feature-rich HTTP router for Go. A viable alternative to [gorilla/mux](http://github.com/gorilla/mux).

## Features

- Trie-based router for better performance
- Supports required, optional, regexp and wildcard slots
- Smart slot delimiters (e.g. can match `/{from}-{to}`)
- Well-tested with 100s of tests

## Install

```sh
go get github.com/livebud/mux
```

## Example

```go
package main

import (
	"net/http"

	"github.com/livebud/mux"
)

func main() {
  handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(r.URL.Path))
  })
  router := mux.New()
  router.Get("/", handler)
  router.Get("/users/{id}", handler)
  router.Post("/users/{id}.{format}", handler)
  router.Get("/posts/{post_id}/comments/{id}", handler)
  router.Get("/fly/{from}-{to}", handler)
  router.Get("/v{major|[0-9]+}.{minor|[0-9]+}", handler)
  router.Get("/{owner}/{repo}/{branch}/{path*}", handler)
  http.ListenAndServe(":3000", router)
}
```

## Contributors

- Matt Mueller ([@mattmueller](https://twitter.com/mattmueller))

## License

MIT
