.PHONY: default
default: build

setup-all:
	@export GOPATH=$HOME/workspace
	@export PATH=$PATH:/usr/local/go/bin
	@export GOROOT=/usr/local/go
	@sudo apt-get -f install upx-ucl

build-small:
	@go build -ldflags="-s -w" -o ./cli/cli ./cli/cli.go
	@upx --brute --best --lzma ./cli/cli

build-linux:
	@GOOS=linux go build -ldflags="-s -w" -o ./cli/cli ./cli/cli.go

build-mac:
	@GOOS=darwin go build -o ./cli/cli cli/cli.go

build-all:
	@sudo apt-get install upx-ucl
	@GOOS=windows go build -ldflags="-s -w" -o ./cli/cli.exe ./cli/cli.go
	@upx --brute --best --lzma ./cli/cli.exe
	@GOOS=linux go build -ldflags="-s -w" -o ./cli/cli ./cli/cli.go
	@GOOS=darwin go build -ldflags="-s -w" -o ./cli/cli ./cli/cli.go
	
build-windows:
	@GOOS=windows go build cli/cli.go
