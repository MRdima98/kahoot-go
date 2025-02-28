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
	"strings"
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
	no_answer          = ""
)

var answered *websocket.Conn

type Player struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Answer string `json:"answer"`
	Score  int    `json:"score"`
}

type Options struct {
	Ans1 string `json:"ans1"`
	Ans2 string `json:"ans2"`
	Ans3 string `json:"ans3"`
	Ans4 string `json:"ans4"`
}

var colors = map[string]string{
	"red_answer":    "bg-kahootRed",
	"blue_answer":   "bg-kahootBlue",
	"green_answer":  "bg-kahootGreen",
	"yellow_answer": "bg-kahootYellow",
}

// Write down what the player is answering
// the game socket will reset the answered and write down the points
// also will clear the current answer and reset the gray
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
			fmt.Println("Error unmarshaling JSON in for loop get:", err)
			return
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
			player = Player{
				Name:   result["name"].(string),
				Status: connected,
				Answer: no_answer,
				Score:  0,
			}
			savePlayerInfo(player, rdb, connected)

			tmpl, err = template.ParseFiles(playerControlsPath)
			if err != nil {
				log.Println(err)
			}

			var tpl bytes.Buffer
			err = tmpl.Execute(&tpl, readOptions(rdb))
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

	answered = conn
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

func readOptions(rdb *redis.Client) Options {
	fakeData := Options{
		Ans1: "batman",
		Ans2: "aquaman",
		Ans3: "joker",
		Ans4: "superman",
	}

	fake, err := json.Marshal(fakeData)
	if err != nil {
		log.Println("Marshal issues", err)
	}

	err = rdb.Set(ctx, "options", fake, 0).Err()
	if err != nil {
		log.Println("Writing options", err)
	}

	tmp, err := rdb.Get(ctx, "options").Result()
	if err != nil {
		log.Println("Reading options", err)
	}

	var options Options

	err = json.Unmarshal([]byte(tmp), &options)
	if err != nil {
		log.Println("UNmarshal err: ", err)
	}

	fmt.Println(options)

	return options
}

func whichAnswer(answer string, rdb *redis.Client, tmpl *template.Template, conn *websocket.Conn) {
	html := `
			<div id="n_answered" hx-swap-oob="innerHTML">
			%d
			</div>
			`
	html = fmt.Sprintf(html, saveNAnswered(rdb))

	if answered == nil {
		fmt.Println("There is no open game")
		return
	}

	if err := answered.WriteMessage(websocket.TextMessage, []byte(html)); err != nil {
		fmt.Println("Can't sign that a player wrote a message", err)
		return
	}

	var ans_button bytes.Buffer
	err := tmpl.ExecuteTemplate(&ans_button, answer, readOptions(rdb))
	if err != nil {
		log.Println(err)
	}

	gray := strings.ReplaceAll(ans_button.String(), colors[answer], "bg-gray-200")

	fmt.Println(gray)

	fmt.Println("We are doing", gray)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(gray)); err != nil {
		log.Println(err)
		return
	}
}
