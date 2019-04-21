module github.com/go-joe/joe/examples/useful

go 1.12

require (
	github.com/go-joe/joe v0.7.1-0.20190421103757-34b695fe0a2f
	github.com/go-joe/redis-memory v0.3.0
	github.com/go-joe/slack-adapter v0.6.0
	github.com/pkg/errors v0.8.1
)

replace github.com/go-joe/joe => ../..
