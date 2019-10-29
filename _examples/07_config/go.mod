module github.com/go-joe/joe/examples/config

go 1.12

require (
	github.com/go-joe/http-server v0.5.0
	github.com/go-joe/joe v0.9.0
	github.com/go-joe/redis-memory v0.3.2
	github.com/go-joe/slack-adapter v0.7.0
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/pkg/errors v0.8.1
)

replace github.com/go-joe/joe => ../..
