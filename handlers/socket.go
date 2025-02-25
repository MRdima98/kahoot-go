package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var players = []string{}

const (
	playerControlsPath = "templates/playerControls.html"
	connected          = "connected"
	disconnected       = "disconnected"
	playersKey         = "players"
)

type player struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Score  int    `json:"score"`
}

func SocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("conn: ", conn)

	rdb := redisClient()
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

		if result["name"] != nil {
			player := player{
				Name:   result["name"].(string),
				Status: connected,
				Score:  0,
			}
			fmt.Println("player obj: ", player)

			playerJSON, err := json.Marshal(player)
			if err != nil {
				panic(err)
			}

			fmt.Println("My JSON: ", playerJSON)

			err = rdb.Set(ctx, player.Name, playerJSON, 0).Err()
			if err != nil {
				log.Println(err)
			}
		}

		val, err := rdb.Get(ctx, "John").Result()
		if err != nil {
			log.Println(err)
		}
		fmt.Println("John val:", val)

		var player map[string]any
		err = json.Unmarshal([]byte(val), &player)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return
		}
		fmt.Println("John", player)

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
		Addr:     "localhost:6379",
		Password: "",
		DB:       0, // use default DB
	})

	return rdb
}
