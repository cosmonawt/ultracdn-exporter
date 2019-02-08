package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const apiURL = "https://api.leasewebultracdn.com"

type client struct {
	customerID string
	apiToken   string `json:"access_token"`
}

func (c *client) login(username, password string) error {
	if username == "" {
		log.Fatal("no username provided")
	}
	if password == "" {
		log.Fatal("no password provided")
	}

	form := url.Values{}
	form.Add("username", username)
	form.Add("password", password)
	form.Add("grant_type", "password")

	req, err := http.NewRequest(http.MethodPost, apiURL+"/auth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("non 2xx status: %d", res.StatusCode)
	}

	if err = json.NewDecoder(res.Body).Decode(&c); err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}
	return nil
}

func (c *client) getCustomerID() (string, error) {
	req, err := http.NewRequest(http.MethodGet, apiURL+"/self", nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.apiToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non 2xx status: %d", res.StatusCode)
	}

	s := struct {
		Response struct {
			CustomerID string `json:"customerId"`
		} `json:"response"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&s); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	c.customerID = s.Response.CustomerID
	return c.customerID, nil
}

type distributionGroup struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Domain string `json:"domain"`
}

func (c *client) getDistributionGroups(customerID string) ([]distributionGroup, error) {
	req, err := http.NewRequest(http.MethodGet, apiURL+"/"+customerID+"/config/distributiongroups", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.apiToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non 2xx status: %d", res.StatusCode)
	}

	s := struct {
		Response []distributionGroup `json:"response"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&s); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return s.Response, nil
}

type metric struct {
	GroupID string
	Target  string  `json:"target"`
	Points  []point `json:"points"`
}

type point struct {
	Value     float64 `json:"value"`
	Timestamp int     `json:"timestamp"`
}

func (c *client) gatherMetrics(distributionGroupID, metricName string) (metric, error) {
	form := url.Values{}
	form.Add("start", "-30min")
	form.Add("end", "-20min") // Leaseweb aggregates in 5 minute intervals, to make sure we dont scrape 0, we have a lag of 20 minutes.
	form.Add("target", fmt.Sprintf("alias(aggregate(sum(%s.*.*.*.%s),'5min', 'sum', 'true'), '%[2]s')", distributionGroupID, metricName))

	req, err := http.NewRequest(http.MethodPost, apiURL+"/"+c.customerID+"/query", strings.NewReader(form.Encode()))
	if err != nil {
		return metric{}, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.apiToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return metric{}, fmt.Errorf("error making request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return metric{}, fmt.Errorf("non 2xx status: %d", res.StatusCode)
	}

	mr := struct {
		Response []metric `json:"response"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&mr); err != nil {
		return metric{}, fmt.Errorf("error decoding response: %v", err)
	}

	return mr.Response[0], nil
}
