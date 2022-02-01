module github.com/muir/xm

go 1.16

replace github.com/muir/rest v0.0.0 => ../rest

require (
	github.com/gorilla/mux v1.8.0
	github.com/muir/rest v0.0.0
	go.uber.org/zap v1.19.0
)
