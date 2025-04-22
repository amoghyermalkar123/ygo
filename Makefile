.PHONY: all replay build clean run test

test:
	go test -v ./...
	go test -v ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

clean:
	rm tmp/*
	rm -f coverage.out