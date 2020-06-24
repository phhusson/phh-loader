export PATH:=$(PATH):/opt/gcc-linaro-7.5.0-2019.12-x86_64_aarch64-linux-gnu/bin/
all: boot.arm64

boot.arm64: boot.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc  go build -ldflags='-extldflags=-static' -o $@ $^
