.PHONY: default
default: build

build-small:
	@sudo apt-get install upx-ucl
	@go generate ./bin
	@go build -ldflags="-s -w" -o ./bin/run ./bin/server.go
	@upx --brute --best --lzma ./bin/run

setup:
	@export GOPATH=$HOME/workspace
	@export PATH=$PATH:/usr/local/go/bin
	@export GOROOT=/usr/local/go
	
build-windows:
	@GOOS=windows go build test/cli.go

build-linux:
	@go build test/cli.go

build-mac:
	@GOOS=darwin go build test/cli.go

build-all:
	@sudo apt-get install upx-ucl
	@GOOS=windows go build -ldflags="-s -w" -o ./test/cli.exe ./test/cli.go
	@upx --brute --best --lzma ./bin/decrypt-windows.exe
	@GOOS=linux go build -ldflags="-s -w" -o ./test/cli ./test/cli.go
	@GOOS=darwin go build -ldflags="-s -w" -o ./test/cli ./test/cli.go
	
