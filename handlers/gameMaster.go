package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

var Answered = 0

const (
	index      = "index.html"
	indexPath  = "templates/index.html"
	headPath   = "templates/head.html"
	footerPath = "templates/footer.html"
)

var tmpl = template.Must(
	template.ParseFiles(
		indexPath, footerPath, headPath,
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
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

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
			fmt.Println("Error unmarshaling JSON in for loop get:", err)
			return
		}

		fmt.Println(result)
		if result["master"] != nil {
			fmt.Println("master:", result["master"])
		}
	}
}

func GameHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\nGame master in the house!")
	rdb := RedisClient()

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

	err = tmpl.ExecuteTemplate(w, index, struct {
		Path     string
		Answered int
		Current  Question
	}{
		r.URL.Path,
		Answered,
		questions[curr_question],
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	rdb.Close()
}
