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
