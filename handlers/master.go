package handlers

import (
	"bytes"
	"context"
	"encoding/json"
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
	gamePath            = "templates/game.html"
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

func LobbyHandler(w http.ResponseWriter, r *http.Request) {
	lobby_cache, err := r.Cookie("lobby_name")
	restart_game := r.Method == http.MethodPost

	// TODO: Delete the lobby everytime you make a new game
	if restart_game {
		delete(lobbies, lobby_cache.Name)
	}

	var lobby_code string
	if err != nil || restart_game {
		lobby_code = GenRandomKey()
		cookie := http.Cookie{
			Name:     "lobby_name",
			Value:    lobby_code,
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, &cookie)
		cookie.Name = "game"
		cookie.Value = "not_started"
		http.SetCookie(w, &cookie)
		restart_game = true
	} else {
		lobby_code = lobby_cache.Value
	}

	tmpl, err := template.ParseFiles(
		gamePath, footerPath, headPath, lobbyPath, questionInterfacePath,
		leaderBoardPath,
	)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	page_to_render := lobby

	questions := getQuestion(lobby_code)
	err = tmpl.ExecuteTemplate(w, page_to_render, struct {
		Path        string
		Link        string
		Lobby       string
		Players     map[string]Player
		RestartGame bool
		Current     question
		Answered    int
	}{
		r.URL.Path,
		"quizaara.mrdima98.dev/player",
		lobby_code,
		lobbies[lobby].players,
		restart_game,
		questions[lobbies[lobby_code].curr_question],
		Answered,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

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

	if entry, ok := lobbies[lobby]; ok {
		entry.master = conn
		lobbies[lobby] = entry
	} else {
		lobbies[lobby] = Game{
			master:  conn,
			players: make(map[string]Player),
		}
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
			log.Println("Can't parse input", err)
		}

		if string(message) == "timeout" {
			LeaderBoard(redis, lobby)
		}

		if result["start-game"] != nil {
			lobby := result["lobby"].(string)
			loadQuestion(lobby)

			if entry, ok := lobbies[lobby]; ok {
				entry.game_started = true
				lobbies[lobby] = entry
			}
		}

		// TODO: we also need to nuke this previous lobby, only save on redis the
		// info and everything else should just be cleared
		if result["refresh-lobby"] != nil {
			delete(lobbies, lobby)
			lobby := result["lobby-input-refresh"].(string)
			lobbies[lobby] = Game{
				master:  conn,
				players: make(map[string]Player),
			}
			log.Println("refreshed", lobbies)
		}
	}
}

// TODO: split this function into two, so that it is evident
// that we are loading the leader board and THEN we load question
// eg.
// leaderBoard()
// loadQuestion()
// loadQuestion should take an input a seconds of timeout, default is 0
// but and I decide when to wait more
func LeaderBoard(redis *redis.Client, lobby string) {
	if entry, ok := lobbies[lobby]; ok {
		entry.leaderboard_phase = true
		lobbies[lobby] = entry
	}

	tmpl, err := template.ParseFiles(leaderBoardPath)
	if err != nil {
		log.Println(err)
	}

	refresh_lobby(redis, lobby)
	var leaderboard bytes.Buffer
	err = tmpl.Execute(&leaderboard, struct {
		Players map[string]Player
	}{
		lobbies[lobby].players,
	})
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
		log.Printf("Can't read \"%s\"", Questions)
	}

	err = json.Unmarshal([]byte(data), &questions)
	if err != nil {
		log.Println("Can't unmarshal them data")
	}

	data, err = redis.Get(context.Background(), answered+lobby).Result()
	if err != nil {
		log.Printf("Can't read \"%s\"", answered+lobby)
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

	if entry, ok := lobbies[lobby]; ok {
		entry.leaderboard_phase = false
		lobbies[lobby] = entry
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

func loadQuestion(lobby string) {
	redis := RedisClient()
	var questions []question

	rawQuestions, err := redis.Get(context.Background(), Questions).Result()
	if err != nil {
		log.Printf("Can't read \"%s\"", Questions)
	}

	err = json.Unmarshal([]byte(rawQuestions), &questions)
	if err != nil {
		log.Println("Can't unmarshal them data")
	}

	if len(questions) <= lobbies[lobby].curr_question {
		return
	}

	var game_start bytes.Buffer
	err = gameTmpl.ExecuteTemplate(&game_start, "body", struct {
		Answered int
		Current  question
	}{
		lobbies[lobby].answered,
		questions[lobbies[lobby].curr_question],
	})

	if err := lobbies[lobby].master.WriteMessage(websocket.TextMessage, game_start.Bytes()); err != nil {
		log.Println(err)
	}
}

func getQuestion(lobby string) []question {
	redis := RedisClient()
	var questions []question

	rawQuestions, err := redis.Get(context.Background(), Questions).Result()
	if err != nil {
		log.Printf("Can't read \"%s\"", Questions)
	}

	err = json.Unmarshal([]byte(rawQuestions), &questions)
	if err != nil {
		log.Println("Can't unmarshal them data")
	}

	return questions
}
