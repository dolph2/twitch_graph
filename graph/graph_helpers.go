package graph

import (
	"fmt"
	"sync"

	"github.com/jmcvetta/neoism"
)

func AddUsersToDB(user string, names <-chan string, next chan string, db *neoism.Database, dir int) error {
	var wg sync.WaitGroup
	var qs []*neoism.CypherQuery
	error_channel := make(chan error, 10000)
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
			var cq neoism.CypherQuery
			if dir == 0 {
				cq = neoism.CypherQuery{
					Statement: `
					MERGE (user:User {name: {self}})
					MERGE (streamer:User {name: {streamer}})
					MERGE (user)-[:FOLLOWS]->(streamer)
					RETURN user, streamer
				`,
					Parameters: neoism.Props{"self": user, "streamer": name},
				}
			} else if dir == 1 {
				cq = neoism.CypherQuery{
					Statement: `
					MERGE (user:User {name: {self}})
					MERGE (streamer:User {name: {streamer}})
					MERGE (user)<-[:FOLLOWS]-(streamer)
					RETURN user, streamer
				`,
					Parameters: neoism.Props{"self": user, "streamer": name},
				}
			}
			qs = append(qs, &cq)
			next <- name

			if count > 1000 {
				var err error
				if err := tx.Query(qs[old_count:count]); err != nil {
					error_channel <- err
				}
				old_count = count
				tx.Commit()
				tx, err = db.Begin([]*neoism.CypherQuery{})
				if err != nil {
					error_channel <- err
				}
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
