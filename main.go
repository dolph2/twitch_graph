package main

import (
	"fmt"

	"github.com/dolph2/twitch_graph/graph"
	"github.com/dolph2/twitch_graph/twitch"
	"github.com/jmcvetta/neoism"
)

func main() {
	concurrency := 50
	sem := make(chan bool, concurrency)
	db, err := neoism.Connect("http://localhost:7474/db/data")
	if err != nil {
		fmt.Println("ERROR1: ", err.Error())
		return
	}
	next := make(chan string, 10000)
	next <- "hughbliss"
	for user := range next {
		sem <- true
		go func(user string) {
			defer func() { <-sem }()
			followed := make(chan string, 10000)
			if err := twitch.GetFollowed(user, followed); err != nil {
				fmt.Println("ERROR2: ", err.Error())
				return
			}
			if err := graph.AddUsersToDB(user, followed, next, db); err != nil {
				fmt.Println("ERROR3: ", err.Error())
			}

			followers := make(chan string, 10000)
			if err := twitch.GetFollowers(user, followers); err != nil {
				fmt.Println("ERROR4: ", err.Error())
				return
			}
			if err := graph.AddUsersToDB2(user, followers, next, db); err != nil {
				fmt.Println("ERROR5: ", err.Error())
			}
		}(user)
	}
}
