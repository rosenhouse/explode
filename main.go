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
	fmt.Fprintf(w, `<html><head>`+
		`<meta http-equiv="Refresh" content="1">`+
		`</head><body>%s</body></html>`, message)
}

type GameStatus struct {
	InstanceCount int `json:"instance_count"`
	Code          string
	Timeout       int
}

func showStatus(w http.ResponseWriter, gameStatus GameStatus) {

	writeHTML(w, `
	<html><head>

<style type="text/css">
section{
font-family: Courier;
text-align: center;
}

section.lives {
font-size: 6em;
width: 49%;
display: inline-block;
}

section.timer {
font-size: 6em;
width: 49%;
display: inline-block;
border: 3px solid black;
height: 2em;
vertical-align: middle;
}

section.code {
text-align: center;
font-size: 8em;
border: 3px solid black;
height: 2em;
vertical-align: middle;
}
.heart {
color: black;
font-weight: bold;
font-size: 2.5em;
vertical-align: middle;
}
</style>
</head><body>
<section class="lives">
<span class="heart">♥</span> `+fmt.Sprintf("%d", gameStatus.InstanceCount)+`
</section>
<section class="timer">
`+fmt.Sprintf("%d", gameStatus.Timeout)+` sec
</section>
</br>
</br>
<section class="code"> `+
		gameStatus.Code+`
</section>

</body></html>
`)
}

func getInfo(w http.ResponseWriter, r *http.Request) {
	response, err := http.Get("http://" + gameServerHost)
	if err != nil {
		log.Printf("error requesting state from game server: %s", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if response.StatusCode == 404 {
		writeHTML(w, "<h1>Game over!</h1>")
		return
	} else {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("error reading response bytes from game-server: %s", err)
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		var gameStatus GameStatus

		err = json.Unmarshal(bodyBytes, &gameStatus)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			writeHTML(w, fmt.Sprintf(
				"Error decoding JSON response from game-server.  Response body was %q, error: %s.",
				bodyBytes, err))
			return
		}

		if gameStatus.Code == "" {
			writeHTML(w, "<h1>The game has not yet started.</h1>")
			return
		}

		showStatus(w, gameStatus)
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
