package twitch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type Channel struct {
	Banner      []byte            `json:"-"`
	Id          int               `json:"_id"`
	Url         string            `json:"url"`
	Mature      bool              `json:"mature"`
	Teams       []string          `json:"-"`
	Status      string            `json:"status"`
	Logo        []byte            `json:"-"`
	Name        string            `json:"name"`
	VideoBanner []byte            `json:"-"`
	DisplayName string            `json:"display_name"`
	Added       time.Time         `json:"created_at"`
	Delay       int               `json:"delay"`
	Game        string            `json:"game"`
	Links       map[string]string `json:"_links"`
	Updated     time.Time         `json:"updated_at"`
	Background  []byte            `json:"background"`
}

type User struct {
	Id          int               `json:"_id"`
	Name        string            `json:"name"`
	Added       time.Time         `json:"created_at"`
	Updated     time.Time         `json:"updated_at"`
	Links       map[string]string `json:"_links"`
	DisplayName string            `json:"display_name"`
	Logo        []byte            `json:"-"`
	Bio         string            `json:"-"`
	Type        interface{}       `json:"-"`
}

type Follows struct {
	Added   time.Time         `json:"created_at"`
	Links   map[string]string `json:"_links"`
	Channel Channel           `json:"channel"`
	User    User              `json:"user"`
}

type FollowResponse struct {
	Links   map[string]string `json:"_links"`
	Follows []Follows         `json:"follows"`
	Total   int               `json:"_total"`
}

func GetFollowers(stream string, followers chan<- string) error {
	concurrency := 1
	sem := make(chan bool, concurrency)
	var wg sync.WaitGroup
	total := 1
	limit := 100

	for count := 0; count < total; count += limit {
		sem <- true
		wg.Add(1)
		go func(count int) {
			defer func() { <-sem; wg.Done() }()
			url := fmt.Sprintf("https://api.twitch.tv/kraken/channels/"+stream+"/follows?limit=%d&offset=%d", limit, count)
			names, err := FollowRequest(url)
			if err != nil {
				count -= limit
				return
			}
			for _, name := range names.Follows {
				followers <- name.User.DisplayName
			}
			total = names.Total
		}(count)
		if total == 1 {
			wg.Wait()
		}
	}
	wg.Wait()
	close(followers)

	return nil
}

func GetFollowed(user string, followed chan<- string) error {
	concurrency := 1
	sem := make(chan bool, concurrency)
	var wg sync.WaitGroup
	total := 1
	limit := 100
	for count := 0; count < total; count += limit {
		sem <- true
		wg.Add(1)
		go func(count int) {
			defer func() { <-sem; wg.Done() }()
			url := fmt.Sprintf("https://api.twitch.tv/kraken/users/"+user+"/follows/channels?limit=%d&offset=%d", limit, count)
			names, err := FollowRequest(url)
			if err != nil {
				count -= limit
				return
			}

			for _, name := range names.Follows {
				followed <- name.Channel.DisplayName
			}
			total = names.Total
		}(count)
		if total == 1 {
			wg.Wait()
		}
	}
	wg.Wait()
	close(followed)

	return nil
}

func FollowRequest(url string) (FollowResponse, error) {
	var user FollowResponse
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return FollowResponse{}, err
	}
	req.Header.Add("Accept", "application/vnd.twitchtv.v3+json")
	resp, err := client.Do(req)
	if err != nil {
		return FollowResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return FollowResponse{}, err
	}
	if err = json.Unmarshal([]byte(body), &user); err != nil {
		return FollowResponse{}, err
	}

	return user, nil
}
