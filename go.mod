module github.com/faisaloncode/ecommerce-crawler

replace github.com/faisaloncode/ecommerce-crawler/crawler => ./crawler

go 1.23.0

toolchain go1.23.8

require (
	github.com/faisaloncode/ecommerce-crawler/crawler v0.0.0-00010101000000-000000000000
	github.com/labstack/echo/v4 v4.13.3
	google.golang.org/grpc v1.72.0
)

require (
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	golang.org/x/time v0.8.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)
