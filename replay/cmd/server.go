package main

import (
	"log"
	"net/http"
	"ygo/replay"
)

func main() {
	mux := http.NewServeMux()
	replay.StartReplayEngine(mux)
	http.ListenAndServe(":8080", mux)
	log.Fatal("Server started on :8080")
}
