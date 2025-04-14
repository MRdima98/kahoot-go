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
	playerControls     = "playerControls.html"
	indexPath          = "templates/index.html"
	lobbyPath          = "templates/lobby.html"
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
		gamePath, footerPath, headPath, playerControlsPath, indexPath, lobbyPath,
	),
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/lobby", lobbyHandler)
	http.HandleFunc("/player", handlers.PlayerHandler)
	http.HandleFunc("/socket", handlers.PlayerSocketHandler)
	http.HandleFunc("/questions", handlers.GameMasterSocketHandler)

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

func lobbyHandler(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("lobby_name")
	restart_game := r.Method == http.MethodPost
	lobby_code := handlers.GenRandomKey()
	if err != nil || restart_game {
		fmt.Println("No cookie in lobby", err)
		socketCookie := http.Cookie{
			Name:     "lobby_name",
			Value:    lobby_code,
			Path:     "/questions",
			MaxAge:   3600,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}
		lobbyCookie := socketCookie
		playerCookie := socketCookie
		lobbyCookie.Path = "/lobby"
		playerCookie.Path = "/player"
		http.SetCookie(w, &lobbyCookie)
		http.SetCookie(w, &socketCookie)
		http.SetCookie(w, &playerCookie)
	}

	if restart_game {
		err = tmpl.ExecuteTemplate(w, "lobby_code", struct {
			Lobby string
		}{
			lobby_code,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err = tmpl.ExecuteTemplate(w, lobby, struct {
		Path  string
		Link  string
		Lobby string
	}{
		r.URL.Path,
		"quizaara.mrdima98.dev/player",
		"",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
