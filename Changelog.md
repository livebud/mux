# 0.8.0 / 2026-11-15

- **BREAKING** `ServeHTTP` returns 405 with an `Allow` header instead of 404 when the path matches under another method. `r.Middleware(next)` still calls `next`.
- **BREAKING** overlapping routes no longer depend on registration order; the more specific route wins (required > optional > wildcard). For example, `/users/1` used to match `/users/{path*}` if it was registered before `/users/{id}`.
- **BREAKING** `Find` on a `Group` router resolves the route relative to the group, like `Get` (`api.Find("GET", "/x")`, not `"/api/x"`)
- HEAD requests are served by GET routes when there's no HEAD route
- OPTIONS requests get a 204 with an `Allow` header unless an OPTIONS route handles them. Add CORS middleware to the router you serve (`r.Use(cors)`), not to a group.
- `Routes()` lists optional and wildcard routes once, in a stable order
- `Find` works with trailing slashes
- a route that fails to register no longer leaves part of itself behind
- bump enroute to v0.3.0

# 0.7.0 / 2026-11-15

- **BREAKING** panic instead of returning an error. This has been confusing
  agents and it's typically a developer error anyway. If you need to return an
  error use `r.Set(http.MethodGet, ...)`

# 0.6.0 / 2026-08-31

- **BREAKING**: you must now do `r = r.Use(mw)`, as returns a new router instance
- **BREAKING**: remove `mux.Routes`
- **BREAKING**: `Mountable` now returns a `*Router`
- add support route-specific middleware
- add support for passing in multiple middleware into `r.Use(mw...)`
- `r.Mount(mnt)` now returns a `*Router`

# 0.5.0 / 2026-02-01

- **BREAKING** rename `mux.Interface` to `mux.Routes`
- add support for mux.Mount(m mux.Mountable)

# 0.4.0 / 2026-01-25

- **BREAKING** switch `Use(mw Middleware)` to accept a `Middleware` interface instead of a function
- Add `mux.Use(fn)` and `mux.Compose(...mw)`.
- Add tests for `subrouter := router.Group(path)`
- sanity check for /.well-known

# 0.3.0 / 2025-08-31

- **BREAKING:** switch back to using `http.Handler` instead of `http.HandlerFunc`. This just ends up being more flexible (e.g. allow for chaining) and it's easy enough to do `http.HandlerFunc(fn)`.

# 0.2.0 / 2025-05-12

- bump enroute to improve matching

# 0.1.1 / 2024-03-16

- switch to using functions
- add middleware stack support back in

# 0.1.0 / 2024-03-16

- run go mod tidy and rename history to changelog
- reduce scope of mux.Router
- move the radix tree logic into enroute

# 0.0.10 / 2024-01-24

- don't apply slots when there are extensions
- fix reading from request body

# 0.0.9 / 2023-10-27

- middleware should also run on non-matching routes

# 0.0.8 / 2023-10-27

- add middleware support
- add batch support
- add layouts and errors

# 0.0.7 / 2023-10-07

- rename mux.List() to mux.Routes()

# 0.0.6 / 2023-09-24

- add .Match(method, path) method

# 0.0.5 / 2023-09-24

- fix panic edge case

# 0.0.4 / 2023-09-04

- add `-race` to makefile
- add `.Find(route)` and `.List()` route methods

# 0.0.3 / 2023-08-20

- fix edge case where wrong handler was getting called

  If you added `/{id}/edit`, then added `/`, the parent handler would get the previous handler. So if you called `http.Get("/")`, it would trigger the edit handler.

# 0.0.2 / 2023-08-20

- expose the AST and add a top-level mux.Parse function
- add staticcheck

# 0.0.1 / 2023-08-14

- initial commit
