package main

import (
	"context"
	"fmt"
	"kahoot/handlers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"text/template"

	"github.com/gorilla/websocket"
)

const (
	index              = "index.html"
	game               = "game.html"
	lobby              = "lobby.html"
	playerMenu         = "playerMenu.html"
	playerControls     = "playerControls.html"
	indexPath          = "templates/index.html"
	lobbyPath          = "templates/lobby.html"
	gamePath           = "templates/game.html"
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
		gamePath, footerPath, headPath, playerControlsPath, playerMenuPath, indexPath, lobbyPath,
	),
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/lobby", lobbyHandler)
	http.HandleFunc("/game", handlers.GameHandler)
	http.HandleFunc("/player", playerHandler)
	http.HandleFunc("/socket", handlers.PlayerHandler)
	http.HandleFunc("/questions", handlers.QuestionsHandler)

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

func playerHandler(w http.ResponseWriter, r *http.Request) {
	// if r.Method == "POST" {
	// 	fmt.Println(r.FormValue("name"))
	// 	err := tmpl.ExecuteTemplate(w, playerControls, nil)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	}
	// 	return
	// }

	queryParams := r.URL.Query()
	sara := false

	for _, values := range queryParams {
		for _, el := range values {
			if el == "Sara" {
				sara = true
			}
		}
	}

	err := tmpl.ExecuteTemplate(w, playerMenu, struct {
		Path string
		Sara bool
	}{
		r.URL.Path,
		sara,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func lobbyHandler(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, lobby, struct {
		Path string
		Link string
	}{
		r.URL.Path,
		"quizaara.mrdima98.dev/player",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
