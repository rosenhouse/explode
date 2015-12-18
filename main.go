package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {

	gameServerHost := "game-server.cfapps.io"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		r.RequestURI = ""
		r.URL.Host = gameServerHost
		r.URL.Scheme = "http"
		r.Host = gameServerHost

		log.Printf("got request: %+v\n", r)

		response, err := http.DefaultClient.Do(r)
		if err != nil {
			log.Printf("error reaching game server: %s", err)
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("error reading response bytes from game-server", err)
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		if response.StatusCode != 404 {
			w.WriteHeader(response.StatusCode)
			w.Write(bodyBytes)
		} else {
			if r.Method == "POST" {
				fmt.Fprintf(w, `{"code":"GAME OVER"}`)
			} else {
				fmt.Fprintf(w, "<html><body>Game Over!</body></html>")
			}
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
