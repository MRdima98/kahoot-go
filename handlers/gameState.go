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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 8192,
}

var ctx = context.Background()

const (
	playerControlsPath = "templates/playerControls.html"
	connected          = "connected"
	disconnected       = "disconnected"
	no_answer          = ""
	Questions          = "questions"
	curr_question_key  = "curr_question"
)

var master *websocket.Conn

// TODO - unify under 1 socket
func PlayerHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("\n\nOpened PLAYER connection!")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	var curr_player player
	rdb := RedisClient()

	conn.SetCloseHandler(func(code int, text string) error {
		if curr_player == (player{}) {
			return errors.New("No player on this connection...somehow")
		}

		savePlayerInfo(curr_player, rdb, disconnected)
		delete(lobby, curr_player.Name)

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
			whichAnswer("red_answer", rdb, tmpl, conn)
		}

		if result["ans2"] != nil {
			whichAnswer("blue_answer", rdb, tmpl, conn)
		}

		if result["ans3"] != nil {
			whichAnswer("green_answer", rdb, tmpl, conn)
		}

		if result["ans4"] != nil {
			whichAnswer("yellow_answer", rdb, tmpl, conn)
		}

		if result["name"] != nil {
			curr_player = player{
				Name:   result["name"].(string),
				Status: connected,
				Answer: no_answer,
				Score:  0,
			}
			savePlayerInfo(curr_player, rdb, connected)
			lobby[curr_player.Name] = conn

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

func RedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6969",
		Password: "",
		DB:       0, // use default DB
	})

	return rdb
}
