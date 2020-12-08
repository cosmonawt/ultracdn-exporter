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
	err = c.getMultiCDNDistributionGroups()
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
	Description string `json:"description"`
	ID          string `json:"id"`
	Domain      string
}

func (c *Client) getMultiCDNDistributionGroups() error {
	req, err := http.NewRequest(http.MethodGet, apiURL+"/configurations/"+c.customerID+"/distributions/multi-cdn/volume", nil)
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
		Response []struct {
			DistributionGroup
			Domains  []string `json:"domains"`
			Endpoint string   `json:"endpoint"`
		} `json:"response"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&s); err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}

	distGroups := make([]DistributionGroup, len(s.Response))
	for i, g := range s.Response {
		var domain string
		if len(g.Domains) < 1 {
			domain = g.Endpoint
		} else {
			domain = g.Domains[0]
		}
		g.DistributionGroup.Domain = domain
		distGroups[i] = g.DistributionGroup
	}
	c.DistGroups = distGroups

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
	// Leaseweb aggregates in 5 minute intervals and makes metrics available with a delay of about 15 minutes.
	// To make sure we dont scrape false 0, we have a delay of 30 minutes.
	form.Add("start", "-30min")
	form.Add("end", "-25min")
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
