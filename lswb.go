package main

import (
	"encoding/json"
	"github.com/prometheus/common/log"
	"net/http"
	"net/url"
	"strings"
)

const apiURL = "https://api.leasewebultracdn.com"

type client struct {
	ApiToken     string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (c *client) login(username, password string) {
	form := url.Values{}
	form.Add("username", username)
	form.Add("password", password)
	form.Add("grant_type", "password")

	req, err := http.NewRequest(http.MethodPost, apiURL+"/auth/token", strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatalf("could not create initial login request: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("could not login: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatalf("could not login, status: %d", res.StatusCode)
	}

	if err = json.NewDecoder(res.Body).Decode(&c); err != nil {
		log.Fatalf("could not login: %v", err)
	}
}
