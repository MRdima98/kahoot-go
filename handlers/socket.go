package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	playerControlsPath = "templates/playerControls.html"
	connected          = "connected"
	disconnected       = "disconnected"
)

type Answered struct {
	answers_conn *websocket.Conn
	count        int
}

var answered Answered

type Player struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Score  int    `json:"score"`
}

func SocketHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n\nOpened PLAYER connection!")
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	var player Player
	rdb := redisClient()

	conn.SetCloseHandler(func(code int, text string) error {
		if player == (Player{}) {
			return errors.New("No player on this connection...somehow")
		}

		savePlayerInfo(player, rdb, disconnected)

		rdb.Close()
		return nil
	})

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var result map[string]any
		err = json.Unmarshal(p, &result)
		if err != nil {
			fmt.Println("Error unmarshaling JSON in for loop get:", err)
			return
		}

		if result["answer1"] != nil {
			fmt.Println("Res: ", result)

			html := `
			<div id="n_answered" hx-swap-oob="innerHTML">
			%d
			</div>
			`
			html = fmt.Sprintf(html, saveNAnswered(rdb))

			if err := answered.answers_conn.WriteMessage(websocket.TextMessage, []byte(html)); err != nil {
				log.Println(err)
				return
			}
		}

		if result["name"] != nil {
			player = Player{
				Name:   result["name"].(string),
				Status: connected,
				Score:  0,
			}
			savePlayerInfo(player, rdb, connected)
		}

		tmpl, err := template.ParseFiles(playerControlsPath)
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
			return
		}
	}
}

var ctx = context.Background()

func redisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6969",
		Password: "",
		DB:       0, // use default DB
	})

	return rdb
}

func savePlayerInfo(player Player, rdb *redis.Client, status string) {
	data, err := rdb.Get(ctx, player.Name).Result()
	if err != nil {
		playerJSON, err := json.Marshal(player)
		if err != nil {
			log.Println("Marshal err: ", err)
		}

		err = rdb.Set(ctx, player.Name, playerJSON, time.Duration(time.Minute*15)).Err()
		if err != nil {
			log.Println(err)
		}

		return
	}

	err = json.Unmarshal([]byte(data), &player)
	if err != nil {
		log.Println("UNmarshal err: ", err)
		return
	}

	player.Status = status

	playerJSON, err := json.Marshal(player)
	if err != nil {
		log.Println("Marshal err: ", err)
		return
	}

	err = rdb.Set(ctx, player.Name, playerJSON, 0).Err()
	if err != nil {
		log.Println("Updating player status:", err)
		return
	}

	log.Printf("Player %s %s", player.Name, player.Status)
}

func QuestionsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n\nOpened GAME connection!")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	rdb := redisClient()
	tmp, err := rdb.Get(ctx, "n_answers").Result()
	if err != nil {
		log.Println("No count value in redis", err)
		return
	}

	count, err := strconv.Atoi(tmp)
	if err != nil {
		log.Println("Converting to int", err)
		return
	}

	answered.answers_conn = conn
	answered.count = count

	conn.SetCloseHandler(func(code int, text string) error {
		rdb.Close()
		return nil
	})
}

func saveNAnswered(rdb *redis.Client) int {
	n_answered := "n_answered"
	tmp, err := rdb.Get(ctx, n_answered).Result()
	if err != nil {
		log.Println("Reading n_answered: ", err)
	}

	count, err := strconv.Atoi(tmp)
	if err != nil {
		log.Println("Converting n_answered: ", err)
	}

	count++

	err = rdb.Set(ctx, n_answered, count, 0).Err()
	if err != nil {
		log.Println("Writing n_answered: ", err)
	}

	return count
}
