package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/template"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const (
	doesnt_expire = 0
)

type player struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Answer string `json:"answer"`
	Score  int    `json:"score"`
}

type options struct {
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

var lobby []*websocket.Conn

func savePlayerInfo(player player, rdb *redis.Client, status string) {
	data, err := rdb.Get(ctx, player.Name).Result()
	if err != nil {
		playerJSON, err := json.Marshal(player)
		if err != nil {
			log.Println("Marshal err: ", err)
		}

		err = rdb.Set(ctx, player.Name, playerJSON, doesnt_expire).Err()
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

	err = rdb.Set(ctx, player.Name, playerJSON, doesnt_expire).Err()
	if err != nil {
		log.Println("Updating player status:", err)
		return
	}

	log.Printf("Player %s %s", player.Name, player.Status)
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

func readQuestion(rdb *redis.Client) Question {
	tmp, err := rdb.Get(ctx, Questions).Result()
	if err != nil {
		log.Println("Reading questions", err)
	}

	var options []Question

	err = json.Unmarshal([]byte(tmp), &options)
	if err != nil {
		log.Println("Unmarshal err: ", err)
	}

	return options[curr_question]
}

func whichAnswer(answer string, rdb *redis.Client, tmpl *template.Template, conn *websocket.Conn) {
	html := `
	<div id="n_answered" hx-swap-oob="innerHTML">
	%d
	</div>
	`
	html = fmt.Sprintf(html, saveNAnswered(rdb))

	if master == nil {
		fmt.Println("There is no open game")
		return
	}

	fmt.Println(html)
	if err := master.WriteMessage(websocket.TextMessage, []byte(html)); err != nil {
		fmt.Println("Can't sign that a player wrote a message", err)
		return
	}

	var ans_button bytes.Buffer
	err := tmpl.ExecuteTemplate(&ans_button, answer, readQuestion(rdb))
	if err != nil {
		log.Println(err)
	}

	gray := strings.ReplaceAll(ans_button.String(), colors[answer], "bg-gray-200")

	if err := conn.WriteMessage(websocket.TextMessage, []byte(gray)); err != nil {
		log.Println(err)
		return
	}
}
