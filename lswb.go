package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const apiURL = "https://api.leasewebultracdn.com"

type Client struct {
	customerID string
	ApiToken   string `json:"access_token"`
	DistGroups []DistributionGroup
}

func (c *Client) Login(username, password string) error {
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

	if err = json.NewDecoder(res.Body).Decode(c); err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}

	err = c.getCustomerID()
	if err != nil {
		return fmt.Errorf("error getting customerID %v", err)
	}
	err = c.getDistributionGroups()
	if err != nil {
		return fmt.Errorf("error getting distributiongroups %v", err)
	}
	return nil
}

func (c *Client) getCustomerID() error {
	req, err := http.NewRequest(http.MethodGet, apiURL+"/self", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.ApiToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("non 2xx status: %d", res.StatusCode)
	}

	s := struct {
		Response struct {
			CustomerID string `json:"customerId"`
		} `json:"response"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&s); err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}

	c.customerID = s.Response.CustomerID
	return nil
}

type DistributionGroup struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Domain string `json:"domain"`
}

func (c *Client) getDistributionGroups() error {
	req, err := http.NewRequest(http.MethodGet, apiURL+"/"+c.customerID+"/config/distributiongroups", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.ApiToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("non 2xx status: %d", res.StatusCode)
	}

	s := struct {
		Response []DistributionGroup `json:"response"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&s); err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}

	c.DistGroups = s.Response
	return nil
}

type Metric struct {
	GroupID string
	Target  string  `json:"target"`
	Points  []Point `json:"points"`
}

type Point struct {
	Value     float64 `json:"value"`
	Timestamp int     `json:"timestamp"`
}

func (c *Client) FetchMetric(distributionGroupID, metricName string) (Metric, error) {
	form := url.Values{}
	form.Add("start", "-30min")
	form.Add("end", "-20min") // Leaseweb aggregates in 5 minute intervals, to make sure we dont scrape 0, we have a lag of 20 minutes.
	form.Add("target", fmt.Sprintf("alias(aggregate(sum(%s.*.*.*.%s),'5min', 'sum', 'true'), '%[2]s')", distributionGroupID, metricName))

	req, err := http.NewRequest(http.MethodPost, apiURL+"/"+c.customerID+"/query", strings.NewReader(form.Encode()))
	if err != nil {
		return Metric{}, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.ApiToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Metric{}, fmt.Errorf("error making request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Metric{}, fmt.Errorf("non 2xx status: %d", res.StatusCode)
	}

	mr := struct {
		Response []Metric `json:"response"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&mr); err != nil {
		return Metric{}, fmt.Errorf("error decoding response: %v", err)
	}

	if len(mr.Response) > 0 {
		return mr.Response[0], nil
	}
	return Metric{}, nil
}
