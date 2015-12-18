package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<html><body>Game Over!</body></html>")
	})

	port := os.Getenv("PORT")
	if port == "" {
		log.Println("no port set, defaulting...")
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
