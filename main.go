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
	"github.com/joho/godotenv"
)

const (
	index              = "index.html"
	game               = "game.html"
	playerControls     = "playerControls.html"
	indexPath          = "templates/index.html"
	gamePath           = "templates/game.html"
	headPath           = "templates/head.html"
	footerPath         = "templates/footer.html"
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
		gamePath, footerPath, headPath, playerControlsPath, indexPath,
	),
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/lobby", handlers.LobbyHandler)
	http.HandleFunc("/player", handlers.PlayerHandler)
	http.HandleFunc("/socket", handlers.PlayerSocketHandler)
	http.HandleFunc("/questions", handlers.GameMasterSocketHandler)

	srv := &http.Server{Addr: ":8001"}

	go func() {
		fmt.Println("Starting service on port 8001")
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
