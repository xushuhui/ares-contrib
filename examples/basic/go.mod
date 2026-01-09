module example

go 1.24.5

replace github.com/xushuhui/ares => /Users/xsh/gp/ares

replace github.com/xushuhui/ares-contrib => /Users/xsh/gp/ares-contrib

require (
	github.com/xushuhui/ares v0.0.0-00010101000000-000000000000
	github.com/xushuhui/ares-contrib v0.0.0-00010101000000-000000000000
)

require (
	github.com/go-chi/chi/v5 v5.2.3 // indirect
	github.com/google/uuid v1.6.0 // indirect
	golang.org/x/time v0.8.0 // indirect
)
