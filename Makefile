.PHONY: all replay build clean run test

replay:
	go run replay/cmd/server.go

test:
	go test -v ./...
	go test -v ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out