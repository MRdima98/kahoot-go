package handlers

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// TODO: just move stuff to the player
const (
	playerControlsPath = "templates/playerControls.html"
	flashCardPath      = "templates/flashcard.html"
	connected          = "connected"
	disconnected       = "disconnected"
	no_answer          = ""
	Questions          = "questions"
	base_score         = 0
	right_answer       = 100
	sara               = "Sara"
	answered           = "answered"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 8192,
}

var ctx = context.Background()

type Game struct {
	master        *websocket.Conn
	players       map[string]Player
	curr_question int
	answered      int
}

// TODO: At this point I think we just cram everything inside the lobbies, and
// periodically keep saving it into memory
var lobbies = make(map[string]Game)
var whichGame string

func RedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6969",
		Password: "",
		DB:       0,
	})

	return rdb
}

// if err != nil {
// 	_, file, line, _ := runtime.Caller(1)
// 	log.Fatalf("Not a numba %s:%d - %v", file, line, err)
// }
