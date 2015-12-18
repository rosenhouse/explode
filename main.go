package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const gameServerHost = "game-server.cfapps.io"

func postProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		panic("expecting POST request only")
	}

	r.RequestURI = ""
	r.URL.Host = gameServerHost
	r.URL.Scheme = "http"
	r.Host = gameServerHost

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
		fmt.Fprintf(w, `{"code":"GAME OVER"}`)
	}
}

func writeHTML(w http.ResponseWriter, message string) {
	fmt.Fprintf(w, `<html><head><meta http-equiv="Refresh" content="5"></head><body>%s</body></html>`, message)
}

func getInfo(w http.ResponseWriter, r *http.Request) {
	response, err := http.Get("http://" + gameServerHost)
	if err != nil {
		log.Printf("error requesting state from game server: %s", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if response.StatusCode == 404 {
		writeHTML(w, "Game over!")
		return
	} else {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("error reading response bytes from game-server: %s", err)
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		var gameStatus struct {
			InstanceCount int `json:"instance_count"`
			Code          string
		}

		err = json.Unmarshal(bodyBytes, &gameStatus)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			writeHTML(w, fmt.Sprintf(
				"Error decoding JSON response from game-server.  Response body was %q, error: %s.",
				bodyBytes, err))
			return
		}

		if gameStatus.InstanceCount < 1 {
			writeHTML(w, "Game server returned an invalid response")
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		writeHTML(w, fmt.Sprintf("There are %d instances<br>The code is %q", gameStatus.InstanceCount, gameStatus.Code))
		return
	}
}

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("got request: %+v\n", r)
		if r.Method == "POST" {
			postProxy(w, r)
		} else {
			getInfo(w, r)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
