package main

import (
	"fmt"
	"sync"

	"github.com/dolph2/twitch_graph/graph"
	"github.com/dolph2/twitch_graph/twitch"
	"github.com/jmcvetta/neoism"
)

func main() {

	concurrency := 1
	sem := make(chan bool, concurrency)
	db, err := neoism.Connect("http://localhost:7474/db/data")
	if err != nil {
		fmt.Println("ERROR1: ", err.Error())
		return
	}
	next := make(chan string, 100000)
	next <- "hughbliss"
	for user := range next {
		sem <- true
		go func(user string) {
			defer func() { <-sem }()
			var followed_wg sync.WaitGroup
			followed := make(chan string, 100000)
			followed_wg.Add(1)
			go func() {
				defer followed_wg.Done()
				fmt.Printf("GETTING %s's followed channels\n", user)
				if err := twitch.GetFollowed(user, followed); err != nil {
					fmt.Println("ERROR2: ", err.Error())
				}
			}()

			fmt.Printf("ADDING %s's followed channels to db\n", user)
			if err := graph.AddUsersToDB(user, followed, next, db, 0); err != nil {
				fmt.Println("ERROR3: ", err.Error())
			}
			followed_wg.Wait()
		}(user)
		sem <- true
		go func(user string) {
			defer func() { <-sem }()
			var follower_wg sync.WaitGroup
			followers := make(chan string, 10000)
			follower_wg.Add(1)
			go func() {
				defer follower_wg.Done()
				fmt.Printf("GETTING %s's followers\n", user)
				if err := twitch.GetFollowers(user, followers); err != nil {
					fmt.Println("ERROR2: ", err.Error())
				}
			}()
			fmt.Printf("ADDING %s's followers to db\n", user)
			if err := graph.AddUsersToDB(user, followers, next, db, 1); err != nil {
				fmt.Println("ERROR3: ", err.Error())
			}
			follower_wg.Wait()
		}(user)
	}

	/*
		test := make(chan string, 10000)
		next := make(chan string, 10000)
		if err := twitch.GetFollowed("hughbliss", test); err != nil {
			fmt.Println("ERROR: ", err.Error())
		}
		fmt.Println("COUNT: ", len(test))
		//	for name := range test {
		//		fmt.Println("FOLLOWED ", name)
		//	}
		if err := graph.AddFollowedToDB("hughbliss", test, next, db); err != nil {
			fmt.Println("ERROR3: ", err.Error())
		}
	*/
}
