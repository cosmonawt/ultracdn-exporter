package main

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/common/log"
	"net/http"
	"net/url"
	"strings"
)

const apiURL = "https://api.leasewebultracdn.com"

type client struct {
	username     string
	password     string
	customerID   string
	ApiToken     string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (c *client) login(username, password string) {
	if username != "" && password != "" {
		c.username = username
		c.password = password
	}
	if c.username == "" {
		log.Fatal("no username provided")
	}
	if c.password == "" {
		log.Fatal("no password provided")
	}

	form := url.Values{}
	form.Add("username", c.username)
	form.Add("password", c.password)
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

func (c *client) getCustomerID() string {
retry:
	req, err := http.NewRequest(http.MethodGet, apiURL+"/self", nil)
	if err != nil {
		log.Fatal("could not get self: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.ApiToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("could not get self: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			log.Fatal("relogin")
			c.login("", "")
			res.Body.Close()
			goto retry // lets hope this wont blow up :D
		}
		log.Fatalf("could not get self, status: %d", res.StatusCode)
	}

	s := struct {
		Response struct {
			CustomerID string `json:"customerId"`
		} `json:"response"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&s); err != nil {
		log.Fatalf("could not get self: %v", err)
	}

	c.customerID = s.Response.CustomerID
	return c.customerID
}

type distributionGroup struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Domain string `json:"domain"`
}

func (c *client) getDistributionGroups(customerID string) []distributionGroup {
retry:
	req, err := http.NewRequest(http.MethodGet, apiURL+"/"+customerID+"/config/distributiongroups", nil)
	if err != nil {
		log.Fatal("could not get distributiongroups: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.ApiToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("could not get distributiongroups: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			log.Fatal("relogin")
			c.login("", "")
			res.Body.Close()
			goto retry // lets hope this wont blow up :D
		}
		log.Fatalf("could not get distributiongroups, status: %d", res.StatusCode)
	}

	s := struct {
		Response []distributionGroup `json:"response"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&s); err != nil {
		log.Fatalf("could not get self: %v", err)
	}

	return s.Response
}

// alias(aggregate(sum(*.*.*.*.bytesdelivered),'1mon', 'sum', 'true'), 'bytesdelivered')
//aggregate(sum(di-m8fg6caj.*.*.*.requestscount), '5min', 'sum', 'true')

type metric struct {
	Target string  `json:"target"`
	Points []point `json:"points"`
}

type point struct {
	Value     float64 `json:"value"`
	Timestamp int     `json:"timestamp"`
}

func (c *client) gatherMetrics(distributionGroupID string) []metric {
	form := url.Values{}
	form.Add("start", "-30min")
	form.Add("end", "-20min") // Leasweb Aggregates in 5minute intervals, to make sure we dont scrape 0, we have a lag of 20 minutes.
	form.Add("target", fmt.Sprintf("alias(aggregate(sum(%s.*.*.*.bytesdelivered),'5min', 'sum', 'true'), 'bytesdelivered')", distributionGroupID))
	form.Add("target", fmt.Sprintf("alias(aggregate(sum(%s.*.*.*.requestscount),'5min', 'sum', 'true'), 'requestscount')", distributionGroupID))
	form.Add("target", fmt.Sprintf("alias(aggregate(sum(%s.*.*.*.bandwidthbps),'5min', 'sum', 'true'), 'bandwidthbps')", distributionGroupID))
	form.Add("target", fmt.Sprintf("alias(aggregate(sum(%s.*.*.*.cachehit_requests),'5min', 'sum', 'true'), 'cachehit_requests')", distributionGroupID))
	form.Add("target", fmt.Sprintf("alias(aggregate(sum(%s.*.*.*.statuscode_2xx_count),'5min', 'sum', 'true'), 'statuscode_2xx_count')", distributionGroupID))
	form.Add("target", fmt.Sprintf("alias(aggregate(sum(%s.*.*.*.statuscode_4xx_count),'5min', 'sum', 'true'), 'statuscode_4xx_count')", distributionGroupID))
	form.Add("target", fmt.Sprintf("alias(aggregate(sum(%s.*.*.*.statuscode_5xx_count),'5min', 'sum', 'true'), 'statuscode_5xx_count')", distributionGroupID))

	req, err := http.NewRequest(http.MethodPost, apiURL+"/"+c.customerID+"/query", strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatalf("could not create query request: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.ApiToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("could not query: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatalf("could not query, status: %d", res.StatusCode)
	}

	mr := struct {
		Response []metric `json:"response"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&mr); err != nil {
		log.Fatalf("could not query: %v", err)
	}

	return mr.Response
}
