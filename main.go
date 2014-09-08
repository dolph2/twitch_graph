package main

import (
	"fmt"

	"github.com/dolph2/twitch_graph/twitch"
	"github.com/jmcvetta/neoism"
)

func main() {
	db, err := neoism.Connect("http://localhost:7474/db/data")
	if err != nil {
		fmt.Println("ERROR: ", err.Error())
		return
	}

	for true {

	}

}
