package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"
)

//go:embed static
var static embed.FS

func main() {
	fSys, err := fs.Sub(static, "static")
	if err != nil {
		log.Fatal(err)
	}
	fs := http.FileServer(http.FS(fSys))
	http.Handle("/", fs)
	http.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		events := make(chan string)
		go func() {
			for {
				events <- fmt.Sprintf("the time is: %v", time.Now())
			}
		}()
		w.Header().Add("Cache-Control", "no-store")
		w.Header().Add("Content-Type", "text/event-stream")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Connection", "keep-alive")

		timeout := time.After(1 * time.Second)

		select {
		case ev := <-events:
			fmt.Fprintf(w, "data: %s\n\n", ev)
		case <-timeout:
			fmt.Fprintf(w, "nothing to sent\n")
		}

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
