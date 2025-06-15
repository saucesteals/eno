COMMIT_HASH := $(shell git rev-parse --short HEAD)

build-win:
	GOOS=windows GOARCH=amd64 go build -o build/eno-$(COMMIT_HASH).exe ./cmd/eno