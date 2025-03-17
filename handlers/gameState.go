package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// TODO: just move stuff to the player

const (
	playerControlsPath = "templates/playerControls.html"
	flashcardPath      = "templates/flashcard.html"
	connected          = "connected"
	disconnected       = "disconnected"
	no_answer          = ""
	Questions          = "questions"
	curr_question_key  = "curr_question"
	base_score         = 0
	right_answer       = 100
	sara               = "Sara"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 8192,
}
var ctx = context.Background()

type Game struct {
	master  *websocket.Conn
	players map[string]Player
}

var lobbies = make(map[string]Game)
var whichGame string

func PlayerHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("\n\nOpened PLAYER connection!")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	var curr_player Player
	rdb := RedisClient()

	conn.SetCloseHandler(func(code int, text string) error {
		if curr_player == (Player{}) {
			return errors.New("No player on this connection...somehow")
		}

		savePlayerInfo(curr_player, rdb, disconnected)
		delete(lobbies[curr_player.Lobby].players, curr_player.Name)

		rdb.Close()
		return nil
	})
	var tmpl *template.Template

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var result map[string]any
		err = json.Unmarshal(p, &result)
		if err != nil {
			log.Println("Error unmarshaling JSON in for loop get:", err)
			return
		}

		if result["player"] != nil {
			log.Println("This is a player")
		}

		if result["ans1"] != nil {
			whichAnswer("red_answer", rdb, tmpl, conn, result["ans1"].(string), curr_player)
		}

		if result["ans2"] != nil {
			whichAnswer("blue_answer", rdb, tmpl, conn, result["ans2"].(string), curr_player)
		}

		if result["ans3"] != nil {
			whichAnswer("green_answer", rdb, tmpl, conn, result["ans3"].(string), curr_player)
		}

		if result["ans4"] != nil {
			whichAnswer("yellow_answer", rdb, tmpl, conn, result["ans4"].(string), curr_player)
		}

		if result["name"] != nil {
			if !nameCheck(conn) {
				continue
			}
			lobby := result["lobby"].(string)
			name := result["name"].(string)

			// if whichGame == sara && result["pwd"].(string) != "wasp" {
			// 	continue
			// }

			curr_player = Player{
				Name:   name,
				Status: connected,
				Answer: no_answer,
				Score:  base_score,
				Lobby:  lobby,
			}

			savePlayerInfo(curr_player, rdb, connected)

			if game, ok := lobbies[lobby]; ok {
				if game.players == nil {
					game.players = make(map[string]Player)
				}

				if pl, ok := game.players[name]; ok {
					pl.conn = conn
					game.players[name] = pl
				}

				lobbies[lobby] = game
			}

			tmpl, err = template.ParseFiles(playerControlsPath)
			if err != nil {
				log.Println(err)
			}

			var tpl bytes.Buffer
			err = tmpl.Execute(&tpl, readQuestion(rdb))
			if err != nil {
				log.Printf("template execution: %s", err)
			}

			if err := conn.WriteMessage(websocket.TextMessage, tpl.Bytes()); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func nameCheck(conn *websocket.Conn) bool {
	tmpl, err := template.ParseFiles(flashcardPath)
	if err != nil {
		log.Println(err)
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, nil)
	if err != nil {
		log.Printf("template execution: %s", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, tpl.Bytes()); err != nil {
		log.Println(err)
		return false
	}

	return true
}

func RedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6969",
		Password: "",
		DB:       0, // use default DB
	})

	return rdb
}
