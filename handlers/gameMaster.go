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

type Question struct {
	Quest   string `json:"question"`
	Ans1    string `json:"answer1"`
	Ans2    string `json:"answer2"`
	Ans3    string `json:"answer3"`
	Ans4    string `json:"answer4"`
	Correct string `json:"correct"`
	Pic     string `json:"path"`
}

func QuestionsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\nGame master in the house!")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		log.Println(err)
		return
	}

	rdb := RedisClient()

	master = conn

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}

		var result map[string]any
		err = json.Unmarshal(message, &result)
		if err != nil {
			fmt.Println("Error unmarshaling JSON in for loop get:", err)
			return
		}

		if result["timeout"] != nil {
			leadBoard(rdb)
		}
	}
}

func GameHandler(w http.ResponseWriter, r *http.Request) {
	rdb := RedisClient()
	URL = r.URL.Path

	var questions []Question

	data, err := rdb.Get(context.Background(), Questions).Result()
	if err != nil {
		log.Println("We can't find them")
	}

	err = json.Unmarshal([]byte(data), &questions)
	if err != nil {
		log.Println("Can't unmarshal them data")
	}

	data, err = rdb.Get(context.Background(), "n_answered").Result()
	if err != nil {
		log.Println("We can't count")
	}

	Answered, err = strconv.Atoi(data)
	if err != nil {
		log.Println("Converting n_answered: ", err)
	}

	err = gameTmpl.ExecuteTemplate(w, index, struct {
		Path     string
		Answered int
		Current  Question
	}{
		r.URL.Path,
		Answered,
		questions[curr_question],
	})

	if err != nil {
		fmt.Println("Err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	rdb.Close()
}

func leadBoard(rdb *redis.Client) {
	tmpl, err := template.ParseFiles(leaderBoardPath)
	if err != nil {
		log.Println(err)
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, nil)
	if err != nil {
		log.Printf("template execution: %s", err)
	}

	if err := master.WriteMessage(websocket.TextMessage, tpl.Bytes()); err != nil {
		log.Println(err)
		return
	}

	err = rdb.Set(context.Background(), "n_answered", 0, 0).Err()
	if err != nil {
		panic("Can't write questions")
	}

	time.Sleep(5 * time.Second)
	fmt.Println("Slept 5")

	var questions []Question

	data, err := rdb.Get(context.Background(), Questions).Result()
	if err != nil {
		log.Println("We can't find them")
	}

	err = json.Unmarshal([]byte(data), &questions)
	if err != nil {
		log.Println("Can't unmarshal them data")
	}

	data, err = rdb.Get(context.Background(), "n_answered").Result()
	if err != nil {
		log.Println("We can't count")
	}

	Answered, err = strconv.Atoi(data)
	if err != nil {
		log.Println("Converting n_answered: ", err)
	}

	fmt.Println("Loaded data")

	err = gameTmpl.ExecuteTemplate(&tpl, "body", struct {
		Path     string
		Answered int
		Current  Question
	}{
		URL,
		Answered,
		questions[curr_question],
	})
	if err != nil {
		log.Printf("template execution: %s", err)
	}

	fmt.Println("Exe templ")

	if err := master.WriteMessage(websocket.TextMessage, tpl.Bytes()); err != nil {
		log.Println(err)
		return
	}

	tmpl, err = template.ParseFiles(playerControlsPath)
	if err != nil {
		log.Println(err)
	}

	tpl.Reset()
	err = tmpl.Execute(&tpl, readQuestion(rdb))
	if err != nil {
		log.Printf("template execution: %s", err)
	}

	// fmt.Println(tpl.String())

	for _, conn := range lobby {
		fmt.Println("am I even loopi?")
		if err := conn.WriteMessage(websocket.TextMessage, tpl.Bytes()); err != nil {
			log.Println(err)
			return
		}
	}

}
