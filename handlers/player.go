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

type Player struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Answer string `json:"answer"`
	Lobby  string `json:"lobby"`
	Score  int    `json:"score"`
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

var server_lobby = make(map[string]*websocket.Conn)
var client_lobby []Player

func savePlayerInfo(player Player, redis *redis.Client, status string) {
	data, err := redis.Get(ctx, player.Name).Result()
	if err != nil {
		playerJSON, err := json.Marshal(player)
		if err != nil {
			log.Println("Marshal err: ", err)
		}

		err = redis.Set(ctx, player.Name, playerJSON, doesnt_expire).Err()
		if err != nil {
			log.Println(err)
		}

		return
	}

	var redis_pl Player
	err = json.Unmarshal([]byte(data), &redis_pl)
	if err != nil {
		log.Println("UNmarshal err: ", err)
		return
	}

	player.Status = status
	player.Score = redis_pl.Score + right_answer

	playerJSON, err := json.Marshal(player)
	if err != nil {
		log.Println("Marshal err: ", err)
		return
	}

	fmt.Println("Player write: ", player)
	err = redis.Set(ctx, player.Name, playerJSON, doesnt_expire).Err()
	if err != nil {
		log.Println("Updating player status:", err)
		return
	}

	log.Printf("Player %s %s", player.Name, player.Status)
}

func saveNAnswered(redis *redis.Client) int {
	n_answered := "n_answered"
	tmp, err := redis.Get(ctx, n_answered).Result()
	if err != nil {
		log.Println("Reading n_answered: ", err)
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

func readQuestion(redis *redis.Client) question {
	tmp, err := redis.Get(ctx, Questions).Result()
	if err != nil {
		log.Println("Reading questions", err)
	}

	var options []question

	err = json.Unmarshal([]byte(tmp), &options)
	if err != nil {
		log.Println("Unmarshal err: ", err)
	}

	curr_question_string, err := redis.Get(ctx, curr_question_key).Result()
	if err != nil {
		log.Println("We can't find them")
	}

	curr_question, err := strconv.Atoi(curr_question_string)
	if err != nil {
		log.Println("Converting n_answered: ", err)
	}

	return options[curr_question]
}

func whichAnswer(answerColor string, redis *redis.Client, tmpl *template.Template, conn *websocket.Conn, answer string, curr_player Player) {
	html := `
	<div id="n_answered" hx-swap-oob="innerHTML">
	%d
	</div>
	`
	answer_count := saveNAnswered(redis)
	html = fmt.Sprintf(html, answer_count)

	if lobbies[curr_player.Lobby] == nil {
		log.Println("There is no open game")
		return
	}

	if err := lobbies[curr_player.Lobby].WriteMessage(websocket.TextMessage, []byte(html)); err != nil {
		log.Println("Can't sign that a player wrote a message", err)
		return
	}

	var ans_button bytes.Buffer
	err := tmpl.ExecuteTemplate(&ans_button, answerColor, readQuestion(redis))
	if err != nil {
		log.Println(err)
	}

	gray := strings.ReplaceAll(ans_button.String(), colors[answerColor], "bg-gray-200")

	if err := conn.WriteMessage(websocket.TextMessage, []byte(gray)); err != nil {
		log.Println(err)
		return
	}

	curr_question_string, err := redis.Get(ctx, curr_question_key).Result()
	if err != nil {
		log.Println("We can't find them")
	}

	curr_question, err := strconv.Atoi(curr_question_string)
	if err != nil {
		log.Println("Converting n_answered: ", err)
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

	if questions[curr_question].Correct == answer {
		fmt.Println("Before func", curr_player)
		savePlayerInfo(curr_player, redis, connected)
	}

	if answer_count == len(server_lobby) {
		LeaderBoard(redis, curr_player.Lobby)
	}
}
