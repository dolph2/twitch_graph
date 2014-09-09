package graph

import (
	//"encoding/json"
	"fmt"
	"sync"

	"github.com/jmcvetta/neoism"
)

func AddFollowedToDB(user string, names <-chan string, next chan<- string, db *neoism.Database) error {
	var wg sync.WaitGroup
	var qs []*neoism.CypherQuery
	error_channel := make(chan error, 100000)
	concurrency := 1
	sem := make(chan bool, concurrency)
	old_count := 0
	count := 0
	fmt.Printf("")

	tx, err := db.Begin([]*neoism.CypherQuery{})
	if err != nil {
		return err
	}
	for name := range names {
		wg.Add(1)
		sem <- true
		go func(name string, count int) {
			defer func() { wg.Done(); <-sem }()
			cq := neoism.CypherQuery{
				Statement: `
					MERGE (user:User {name: {self}})
					MERGE (streamer:User {name: {streamer}})
					MERGE (user)-[:FOLLOWS]->(streamer)
				`,
				Parameters: neoism.Props{"self": user, "streamer": name},
			}
			qs = append(qs, &cq)
			next <- name

			if count > 10000 {
				if err := tx.Query(qs[old_count:count]); err != nil {
					error_channel <- err
				}
				old_count = count
			}
		}(name, count)
		count++
	}
	wg.Wait()
	if old_count != count {
		if err := tx.Query(qs[old_count:count]); err != nil {
			return err
		}
	}
	select {
	case err := <-error_channel:
		return err
	default:
		tx.Commit()
		return nil
	}

}

func AddFollowersToDB(user string, names <-chan string, next chan<- string, db *neoism.Database) error {
	var wg sync.WaitGroup
	var qs []*neoism.CypherQuery
	old_count := 0
	count := 0
	fmt.Printf("")

	tx, err := db.Begin([]*neoism.CypherQuery{})
	if err != nil {
		return err
	}
	for name := range names {
		cq := neoism.CypherQuery{
			Statement: `
					MERGE (user:User {name: {self}})
					MERGE (streamer:User {name: {streamer}})
					MERGE (user)<-[:FOLLOWS]-(streamer)
				`,
			Parameters: neoism.Props{"self": user, "streamer": name},
		}
		qs = append(qs, &cq)
		next <- name

		if count > 1000 {
			if err := tx.Query(qs[old_count:count]); err != nil {
				return err
			}
			old_count = count
		}
		count++
	}
	if old_count != count {
		if err := tx.Query(qs[old_count:count]); err != nil {
			return err
		}
	}
	tx.Commit()

	return nil
}
