.PHONY: all replay build clean run test

replay:
	cd replay && $(MAKE) build
	go run replay/cmd/server.go

test:
	go test -v ./...
	go test -v ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

clean:
	rm tmp/*
	rm -f coverage.out