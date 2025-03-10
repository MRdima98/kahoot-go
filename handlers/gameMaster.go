package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var Answered = 0
var URL string

const (
	index           = "index.html"
	leaderBoard     = "leaderBoard.html"
	leaderBoardPath = "templates/leaderBoard.html"
	indexPath       = "templates/index.html"
	headPath        = "templates/head.html"
	footerPath      = "templates/footer.html"
)

var gameTmpl = template.Must(
	template.ParseFiles(
		indexPath, footerPath, headPath, leaderBoardPath,
	),
)

type question struct {
	Quest   string `json:"question"`
	Ans1    string `json:"answer1"`
	Ans2    string `json:"answer2"`
	Ans3    string `json:"answer3"`
	Ans4    string `json:"answer4"`
	Correct string `json:"correct"`
	Pic     string `json:"path"`
}

func QuestionsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("\nGame master in the house!")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	redis := RedisClient()
	master = conn

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var result map[string]any
		err = json.Unmarshal(message, &result)
		if err != nil {
			log.Println("Error unmarshaling JSON in for loop get:", err)
			return
		}

		if result["timeout"] != nil {
			leadBoard(redis)
		}
	}
}

func GameHandler(w http.ResponseWriter, r *http.Request) {
	redis := RedisClient()
	URL = r.URL.Path

	var questions []question

	data, err := redis.Get(context.Background(), Questions).Result()
	if err != nil {
		log.Println("We can't find them")
	}

	err = json.Unmarshal([]byte(data), &questions)
	if err != nil {
		log.Println("Can't unmarshal them data")
	}

	data, err = redis.Get(context.Background(), "n_answered").Result()
	if err != nil {
		log.Println("We can't count")
	}

	Answered, err = strconv.Atoi(data)
	if err != nil {
		log.Println("Converting n_answered: ", err)
	}

	redis = RedisClient()

	err = redis.Set(context.Background(), curr_question_key, 0, 0).Err()
	if err != nil {
		log.Println("Can't check which quest", err)
	}

	curr_question := 0

	err = gameTmpl.ExecuteTemplate(w, index, struct {
		Path     string
		Answered int
		Current  question
	}{
		r.URL.Path,
		Answered,
		questions[curr_question],
	})

	if err != nil {
		log.Println("Err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	redis.Close()
}

func leadBoard(redis *redis.Client) {
	tmpl, err := template.ParseFiles(leaderBoardPath)
	if err != nil {
		log.Println(err)
	}

	var tpl bytes.Buffer
	fmt.Println("clien lobby", client_lobby)
	err = tmpl.Execute(&tpl, client_lobby)
	if err != nil {
		log.Printf("template execution: %s", err)
	}

	if err := master.WriteMessage(websocket.TextMessage, tpl.Bytes()); err != nil {
		log.Println(err)
		return
	}

	err = redis.Set(context.Background(), "n_answered", 0, 0).Err()
	if err != nil {
		panic("Can't write questions")
	}

	time.Sleep(5 * time.Second)

	var questions []question

	data, err := redis.Get(context.Background(), Questions).Result()
	if err != nil {
		log.Println("We can't find them")
	}

	err = json.Unmarshal([]byte(data), &questions)
	if err != nil {
		log.Println("Can't unmarshal them data")
	}

	data, err = redis.Get(context.Background(), "n_answered").Result()
	if err != nil {
		log.Println("We can't count")
	}

	Answered, err = strconv.Atoi(data)
	if err != nil {
		log.Println("Converting n_answered: ", err)
	}

	curr_question_string, err := redis.Get(context.Background(), curr_question_key).Result()
	if err != nil {
		log.Println("We can't find them")
	}

	curr_question, err := strconv.Atoi(curr_question_string)
	if err != nil {
		log.Println("Converting n_answered: ", err)
	}

	curr_question++
	err = redis.Set(context.Background(), curr_question_key, curr_question, 0).Err()
	if err != nil {
		log.Println("Can't check which quest", err)
	}

	if curr_question == len(questions) {
		return
	}

	err = gameTmpl.ExecuteTemplate(&tpl, "body", struct {
		Path     string
		Answered int
		Current  question
	}{
		URL,
		Answered,
		questions[curr_question],
	})
	if err != nil {
		log.Printf("template execution: %s", err)
	}

	if err := master.WriteMessage(websocket.TextMessage, tpl.Bytes()); err != nil {
		log.Println(err)
		return
	}

	tmpl, err = template.ParseFiles(playerControlsPath)
	if err != nil {
		log.Println(err)
	}

	tpl.Reset()
	err = tmpl.Execute(&tpl, readQuestion(redis))
	if err != nil {
		log.Printf("template execution: %s", err)
	}

	for _, conn := range server_lobby {
		if err := conn.WriteMessage(websocket.TextMessage, tpl.Bytes()); err != nil {
			log.Println(err)
			return
		}
	}

}
