package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var Answered = 0
var URL string

const (
	game                = "game.html"
	leaderBoardTemplate = "leaderBoard.html"
	leaderBoardPath     = "templates/leaderBoard.html"
	gamePath            = "templates/game.html"
	headPath            = "templates/head.html"
	footerPath          = "templates/footer.html"
)

var gameTmpl = template.Must(
	template.ParseFiles(
		gamePath, footerPath, headPath, leaderBoardPath,
	),
)

type question struct {
	Quest   string `json:"question"`
	Ans1    string `json:"answer1"`
	Ans2    string `json:"answer2"`
	Ans3    string `json:"answer3"`
	Ans4    string `json:"answer4"`
	Correct string `json:"correct"`
	Pic     string `json:"path"`
}

// TODO: the master should really not reset on reload, rather keep same lobby
// unless you click "start new game"
// This boils down to check if I have a cookie, if not create one
// For security reasons I should definitely encode them
func GameMasterSocketHandler(w http.ResponseWriter, r *http.Request) {
	lobby := "default value"
	cookie, err := r.Cookie("lobby_name")
	if err != nil {
		log.Printf("%s \"Lobby name\"", err)
	} else {
		lobby = cookie.Value
	}

	log.Printf("Game master in the house! %s", lobby)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	redis := RedisClient()
	lobbyHTML := `<strong id="lobby" hx-swap-oob="outerHTML"> %s </strong>`
	lobby_inputHTML := `<input id="lobby-input" type="text" name="lobby" value="%s" hidden />`
	lobbyHTML = fmt.Sprintf(lobbyHTML, lobby)
	lobby_inputHTML = fmt.Sprintf(lobby_inputHTML, lobby)

	if entry, ok := lobbies[lobby]; ok {
		entry.master = conn
		lobbies[lobby] = entry
	} else {
		lobbies[lobby] = Game{
			master:  conn,
			players: make(map[string]Player),
		}
	}

	// fmt.Println(lobbies)

	if err := conn.WriteMessage(websocket.TextMessage, []byte(lobbyHTML)); err != nil {
		log.Println(err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, []byte(lobby_inputHTML)); err != nil {
		log.Println(err)
		return
	}

	conn.SetCloseHandler(func(code int, text string) error {
		// delete(lobbies, lobby)
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var result map[string]any
		err = json.Unmarshal(message, &result)
		if err != nil {
			log.Println("Error unmarshaling JSON in for loop get:", err)
			return
		}

		if result["timeout"] != nil {
			LeaderBoard(redis, result["lobby"].(string))
		}

		if result["start-game"] != nil {
			fmt.Println("Lobbies: ", lobbies)
			loadFirstQuestion(result["lobby"].(string))
		}
	}
}

// TODO: Leaderboard should be refactored
func LeaderBoard(redis *redis.Client, lobby string) {
	tmpl, err := template.ParseFiles(leaderBoardPath)
	if err != nil {
		log.Println(err)
	}

	refresh_lobby(redis, lobby)
	var leaderboard bytes.Buffer
	err = tmpl.Execute(&leaderboard, lobbies[lobby].players)
	if err != nil {
		log.Printf("template execution: %s", err)
	}

	if err := lobbies[lobby].master.WriteMessage(websocket.TextMessage, leaderboard.Bytes()); err != nil {
		log.Println(err)
		return
	}

	err = redis.Set(context.Background(), answered+lobby, 0, 0).Err()
	if err != nil {
		panic("Can't write questions")
	}

	time.Sleep(5 * time.Second)

	var questions []question

	data, err := redis.Get(context.Background(), Questions).Result()
	if err != nil {
		log.Println("We can't find them")
	}

	err = json.Unmarshal([]byte(data), &questions)
	if err != nil {
		log.Println("Can't unmarshal them data")
	}

	data, err = redis.Get(context.Background(), answered+lobby).Result()
	if err != nil {
		log.Println("We can't count")
	}

	Answered, err = strconv.Atoi(data)
	if err != nil {
		log.Println("Converting n_answered: ", err)
	}

	if entry, ok := lobbies[lobby]; ok {
		entry.curr_question++
		lobbies[lobby] = entry
	}

	if lobbies[lobby].curr_question == len(questions) {
		return
	}

	err = gameTmpl.ExecuteTemplate(&leaderboard, "body", struct {
		Path     string
		Answered int
		Current  question
	}{
		URL,
		Answered,
		questions[lobbies[lobby].curr_question],
	})
	if err != nil {
		log.Printf("template execution: %s", err)
	}

	if err := lobbies[lobby].master.WriteMessage(websocket.TextMessage, leaderboard.Bytes()); err != nil {
		log.Println(err)
		return
	}

	tmpl, err = template.ParseFiles(playerControlsPath)
	if err != nil {
		log.Println(err)
	}

	leaderboard.Reset()
	err = tmpl.Execute(&leaderboard, readQuestion(redis, lobby))
	if err != nil {
		log.Printf("template execution: %s", err)
	}

	for _, pl := range lobbies[lobby].players {
		if err := pl.conn.WriteMessage(websocket.TextMessage, leaderboard.Bytes()); err != nil {
			log.Println(err)
			return
		}
	}

}

func refresh_lobby(redis *redis.Client, lobby string) {
	for key, pl := range lobbies[lobby].players {
		var player Player
		data, err := redis.Get(ctx, pl.Name).Result()
		if err != nil {
			log.Printf("Can't find player: %s", err)
		}

		err = json.Unmarshal([]byte(data), &player)
		if err != nil {
			log.Println("UNmarshal err: ", err)
			return
		}

		if game, ok := lobbies[lobby]; ok {
			if entry, ok := game.players[key]; ok {
				entry.Score = player.Score
				game.players[key] = entry
			}
			lobbies[lobby] = game
		}
	}
}

func loadFirstQuestion(lobby string) {
	redis := RedisClient()
	var questions []question

	data, err := redis.Get(context.Background(), Questions).Result()
	if err != nil {
		log.Println("We can't find them")
	}

	err = json.Unmarshal([]byte(data), &questions)
	if err != nil {
		log.Println("Can't unmarshal them data")
	}

	// TODO: Move counter to game object
	data, err = redis.Get(context.Background(), answered+lobby).Result()
	if err != nil {
		log.Println("We can't count")
	}

	Answered, err = strconv.Atoi(data)
	if err != nil {
		log.Println("Converting n_answered: ", err)
	}

	redis = RedisClient()

	if entry, ok := lobbies[lobby]; ok {
		entry.curr_question = 0
		lobbies[lobby] = entry
	}

	var game_start bytes.Buffer
	err = gameTmpl.ExecuteTemplate(&game_start, "body", struct {
		Answered int
		Current  question
	}{
		Answered,
		questions[lobbies[lobby].curr_question],
	})

	if err := lobbies[lobby].master.WriteMessage(websocket.TextMessage, game_start.Bytes()); err != nil {
		log.Println(err)
		return
	}
}
