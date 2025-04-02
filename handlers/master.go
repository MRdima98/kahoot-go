package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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
func GameMasterSocketHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("\nGame master in the house!")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	redis := RedisClient()
	lobby := genRandomLobby()
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

	if err := conn.WriteMessage(websocket.TextMessage, []byte(lobbyHTML)); err != nil {
		log.Println(err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, []byte(lobby_inputHTML)); err != nil {
		log.Println(err)
		return
	}

	conn.SetCloseHandler(func(code int, text string) error {
		delete(lobbies, lobby)
		return nil
	})

	// fmt.Println("Lobby! : ", lobbies)

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

// TODO: You should check if the code is already in use
func genRandomLobby() string {
	const alfanumeric = "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	lobby := ""
	const max_range = len(alfanumeric)

	for range 4 {
		rand.New(rand.NewSource(time.Now().Unix()))
		i := rand.Intn(max_range)
		lobby = lobby + string(alfanumeric[i])
	}

	return lobby
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

	fmt.Println("Master lobby: ", lobby)
	fmt.Println("Master curr question: ", lobbies[lobby].curr_question)

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
	for k, pl := range lobbies[lobby].players {
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
			if entry, ok := game.players[k]; ok {
				entry.Score = player.Score
			}
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
