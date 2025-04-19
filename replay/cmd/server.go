package main

import (
	"log"
	"net/http"
	"ygo/replay"
)

func main() {
	mux := http.NewServeMux()
	replay.StartReplayEngine(mux)

	log.Fatal(http.ListenAndServe(":8080", mux), "Server started on :8080")
}
