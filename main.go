package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"text/template"

	"github.com/gorilla/websocket"
)

const (
	index              = "index.html"
	playerMenu         = "playerMenu.html"
	playerControls     = "playerControls.html"
	indexPath          = "templates/index.html"
	headPath           = "templates/head.html"
	footerPath         = "templates/footer.html"
	playerMenuPath     = "templates/playerMenu.html"
	playerControlsPath = "templates/playerControls.html"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var players = []string{}

var Answered = 0
var tmpl = template.Must(
	template.ParseFiles(
		indexPath, footerPath, headPath, playerControlsPath, playerMenuPath,
	),
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/game", gameHandler)
	http.HandleFunc("/player", playerHandler)
	http.HandleFunc("/socket", socketHandler)

	srv := &http.Server{Addr: ":8080"}

	go func() {
		fmt.Println("Starting service on port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")
	srv.Shutdown(context.Background())
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, index, struct {
		Path string
	}{
		r.URL.Path,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, index, struct {
		Path     string
		Answered int
	}{
		r.URL.Path,
		Answered,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func playerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		fmt.Println(r.FormValue("name"))
		err := tmpl.ExecuteTemplate(w, playerControls, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err := tmpl.ExecuteTemplate(w, playerMenu, struct {
		Path string
	}{
		r.URL.Path,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("We hit it")

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var result map[string]any
		err = json.Unmarshal(p, &result)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return
		}

		fmt.Println(result["name"])

		if !slices.Contains(players, result["name"].(string)) {
			players = append(players, result["name"].(string))
		}

		fmt.Println(players)

		tmpl, err := template.ParseFiles(playerControlsPath)
		if err != nil {
			log.Println(err)
		}

		var tpl bytes.Buffer
		err = tmpl.Execute(&tpl, nil)
		if err != nil {
			log.Fatalf("template execution: %s", err)
		}

		if err := conn.WriteMessage(websocket.TextMessage, tpl.Bytes()); err != nil {
			log.Println(err)
			return
		}
	}
}
