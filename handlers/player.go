package handlers

import (
	"bytes"
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

const (
	doesnt_expire = 0
	invalid       = true
	valid         = false
	empty_name    = ""
)

type Player struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Answer string `json:"answer"`
	Lobby  string `json:"lobby"`
	Score  int    `json:"score"`
	conn   *websocket.Conn
}

type player_options struct {
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

// TODO: Players should not have to type again to reconnect on reload, I will
// take care with cookies or mc address
func PlayerHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("\n\nOpened PLAYER connection!")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	var tmpl *template.Template
	var curr_player Player
	redis := RedisClient()

	conn.SetCloseHandler(func(code int, text string) error {
		if curr_player == (Player{}) {
			return errors.New("No player on this connection...somehow")
		}

		// TODO: This should allow the recconect, atm I don't feel like it.
		// savePlayer(curr_player, redis)
		delete(lobbies[curr_player.Lobby].players, curr_player.Name)

		redis.Close()
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
			log.Println("Player handler unmarshall error", err)
			return
		}

		switch {
		case result["ans1"] != nil:
			checkAnswer("red_answer", redis, result["ans1"].(string), curr_player)
		case result["ans2"] != nil:
			checkAnswer("blue_answer", redis, result["ans2"].(string), curr_player)
		case result["ans3"] != nil:
			checkAnswer("green_answer", redis, result["ans3"].(string), curr_player)
		case result["ans4"] != nil:
			checkAnswer("yellow_answer", redis, result["ans4"].(string), curr_player)
		}

		if result["name"] != nil {
			lobby := result["lobby"].(string)
			name := result["name"].(string)

			if invalidName(conn, name) {
				continue
			}

			// if whichGame == sara && result["pwd"].(string) != "wasp" {
			// 	continue
			// }

			curr_player = Player{
				Name:   name,
				Status: connected,
				Answer: no_answer,
				Score:  base_score,
				Lobby:  lobby,
				conn:   conn,
			}

			savePlayerRedis(curr_player, redis)
			savePlayer(curr_player)

			tmpl, err = template.ParseFiles(playerControlsPath)
			if err != nil {
				log.Println(err)
			}

			var tpl bytes.Buffer
			err = tmpl.Execute(&tpl, readQuestion(redis, curr_player.Lobby))
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

func savePlayer(new_player Player) {
	game, ok := lobbies[new_player.Lobby]
	if !ok {
		log.Println("Game object not initialized")
		return
	}

	if player, ok := game.players[new_player.Name]; ok {
		player.conn = new_player.conn
		game.players[new_player.Name] = player
	} else {
		game.players[new_player.Name] = new_player
	}
}

func invalidName(conn *websocket.Conn, name string) bool {
	if name != empty_name {
		return valid
	}

	tmpl, err := template.ParseFiles(flashCardPath)
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

	return invalid
}

func savePlayerRedis(player Player, redis *redis.Client) {
	playerJSON, err := json.Marshal(player)
	if err != nil {
		log.Println("Marshal err: ", err)
	}

	err = redis.Set(ctx, player.Name, playerJSON, doesnt_expire).Err()
	if err != nil {
		log.Println(err)
	}
}

func updatePlayerScore(player Player, redis *redis.Client) {
	data, err := redis.Get(ctx, player.Name).Result()
	if err != nil {
		log.Printf("Can't update the score of this player: %s\n", player.Name)
		return
	}

	var redis_pl Player
	err = json.Unmarshal([]byte(data), &redis_pl)
	if err != nil {
		log.Println("UNmarshal err: ", err)
		return
	}

	player.Score = redis_pl.Score + right_answer

	playerJSON, err := json.Marshal(player)
	if err != nil {
		log.Println("Marshal err: ", err)
		return
	}

	err = redis.Set(ctx, player.Name, playerJSON, doesnt_expire).Err()
	if err != nil {
		log.Println("Updating player status:", err)
		return
	}
}

// TODO: Move this to the game object
func saveNAnswered(redis *redis.Client, lobby string) int {
	n_answered := answered + lobby
	tmp, err := redis.Get(ctx, n_answered).Result()
	if err != nil {
		err = redis.Set(ctx, n_answered, 0, 1*time.Hour).Err()
		if err != nil {
			log.Println("Writing n_answered: ", err)
		}
		tmp = "0"
	}

	count, err := strconv.Atoi(tmp)
	if err != nil {
		log.Println("Converting n_answered: ", err)
	}

	count++

	err = redis.Set(ctx, n_answered, count, 0).Err()
	if err != nil {
		log.Println("Writing n_answered: ", err)
	}

	return count
}

func readQuestion(redis *redis.Client, lobby string) question {
	tmp, err := redis.Get(ctx, Questions).Result()
	if err != nil {
		log.Println("Reading questions", err)
	}

	var options []question

	err = json.Unmarshal([]byte(tmp), &options)
	if err != nil {
		log.Println("Unmarshal err: ", err)
	}

	return options[lobbies[lobby].curr_question]
}

func checkAnswer(answerColor string, redis *redis.Client, answer string, curr_player Player) {
	tmpl, err := template.ParseFiles(playerControlsPath)
	if err != nil {
		log.Println(err)
	}

	var ans_button bytes.Buffer
	answer_count := saveNAnswered(redis, curr_player.Lobby)
	html := `
	<div id="n_answered" hx-swap-oob="innerHTML">
	%d
	</div>
	`
	html = fmt.Sprintf(html, answer_count)
	master := lobbies[curr_player.Lobby].master

	if master == nil {
		log.Println("There is no open game (no master)")
		return
	}

	if err := master.WriteMessage(websocket.TextMessage, []byte(html)); err != nil {
		log.Println("Can't tell the master a player answered", err)
		return
	}

	err = tmpl.ExecuteTemplate(&ans_button, answerColor, readQuestion(redis, curr_player.Lobby))
	if err != nil {
		log.Println("Can't read question for player", err)
	}

	grayButton := strings.ReplaceAll(ans_button.String(), colors[answerColor], "bg-gray-200")

	if err := curr_player.conn.WriteMessage(websocket.TextMessage, []byte(grayButton)); err != nil {
		log.Println("I can't gray the button for a player", err)
		return
	}

	var questions []question

	data, err := redis.Get(ctx, Questions).Result()
	if err != nil {
		log.Println("We can't find them")
	}

	err = json.Unmarshal([]byte(data), &questions)
	if err != nil {
		log.Println("Can't unmarshal them data")
	}

	if questions[lobbies[curr_player.Lobby].curr_question].Correct == answer {
		updatePlayerScore(curr_player, redis)
	}

	if answer_count == len(lobbies[curr_player.Lobby].players) {
		LeaderBoard(redis, curr_player.Lobby)
	}
}
