package graph

import (
	//"encoding/json"
	"fmt"
	"sync"

	"github.com/jmcvetta/neoism"
)

func AddUsersToDB(user string, names <-chan string, next chan<- string, db *neoism.Database) error {
	var wg sync.WaitGroup
	var qs []*neoism.CypherQuery
	old_count := 0
	count := 0
	fmt.Printf("")

	tx, err := db.Begin([]*neoism.CypherQuery{})
	if err != nil {
		fmt.Println("A")
		return err
	}
	for name := range names {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			cq := neoism.CypherQuery{
				Statement: `
					MERGE (user:User {name: {self}})
					MERGE (streamer:User {name: {streamer}})
					MERGE (user)-[:FOLLOWS]->(streamer)
					RETURN user
				`,
				Parameters: neoism.Props{"self": user, "streamer": name},
			}
			qs = append(qs, &cq)
			next <- name
		}(name)
		if count > 1000 {
			if err := tx.Query(qs[old_count:count]); err != nil {
				fmt.Println("B")
				return err
			}
			old_count = count
		}
		count++
	}
	wg.Wait()
	if old_count != count {
		if err := tx.Query(qs[old_count:count]); err != nil {
			fmt.Println("C")
			return err
		}
	}
	tx.Commit()

	return nil
}

func AddUsersToDB2(user string, names <-chan string, next chan<- string, db *neoism.Database) error {
	var wg sync.WaitGroup
	var qs []*neoism.CypherQuery
	old_count := 0
	count := 0
	fmt.Printf("")

	tx, err := db.Begin([]*neoism.CypherQuery{})
	if err != nil {
		fmt.Println("A")
		return err
	}
	for name := range names {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			cq := neoism.CypherQuery{
				Statement: `
					MERGE (user:User {name: {self}})
					MERGE (streamer:User {name: {streamer}})
					MERGE (streamer)-[:FOLLOWS]->(user)
					RETURN user
				`,
				Parameters: neoism.Props{"self": user, "streamer": name},
			}
			qs = append(qs, &cq)
			next <- name
		}(name)
		if count > 1000 {
			if err := tx.Query(qs[old_count:count]); err != nil {
				fmt.Println("B")
				return err
			}
			old_count = count
		}
		count++
	}
	wg.Wait()
	if old_count != count {
		if err := tx.Query(qs[old_count:count]); err != nil {
			fmt.Println("C")
			return err
		}
	}
	tx.Commit()

	return nil
}
