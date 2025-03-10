package main

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
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

func main() {
	LoadQuestions()
}

func LoadQuestions() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6969",
		Password: "",
		DB:       0, // use default DB
	})

	data, err := os.ReadFile("./db/db.json")
	if err != nil {
		panic("Can't read json file")
	}

	err = rdb.Set(context.Background(), "questions", data, 0).Err()
	if err != nil {
		panic("Can't write questions")
	}
}
