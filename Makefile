.PHONY: build
build:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64  go build -ldflags="-s -w" -o libs/libp2p-proxy-arm64.dylib  -buildmode=c-shared .
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64  go build -ldflags="-s -w" -o libs/libp2p-proxy-amd64.dylib  -buildmode=c-shared .
	CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o libs/libp2p-proxy.dll -buildmode=c-shared .
	go build -ldflags="-s -w" -o libs/libp2p-proxy .