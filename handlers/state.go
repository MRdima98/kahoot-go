package handlers

import (
	"context"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const (
	playerControlsPath    = "templates/playerControls.html"
	flashCardPath         = "templates/flashcard.html"
	connected             = "connected"
	disconnected          = "disconnected"
	no_answer             = ""
	Questions             = "questions"
	base_score            = 0
	first_question        = 0
	right_answer          = 100
	sara                  = "Sara"
	answered              = "answered"
	headPath              = "templates/head.html"
	footerPath            = "templates/footer.html"
	playerFormPath        = "templates/playerForm.html"
	lobbyPath             = "templates/lobby.html"
	questionInterfacePath = "templates/questionInterface.html"
	lobby                 = "lobby.html"
	leaderBoardPath       = "templates/leaderBoard.html"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 8192,
}

var ctx = context.Background()

type Game struct {
	master            *websocket.Conn
	players           map[string]Player
	curr_question     int
	answered          int
	game_started      bool
	leaderboard_phase bool
}

// TODO: At this point I think we just cram everything inside the lobbies, and
// periodically keep saving it into memory
var lobbies = make(map[string]Game)
var whichGame string

// TODO: this should depened on enviromental variable
func RedisClient() *redis.Client {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "host.docker.internal:6379",
		Password: "",
		DB:       0,
	})

	return rdb
}

// if err != nil {
// 	_, file, line, _ := runtime.Caller(1)
// 	log.Fatalf("Not a numba %s:%d - %v", file, line, err)
// }

// TODO: You should check if the code is already in use
func GenRandomKey() string {
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
