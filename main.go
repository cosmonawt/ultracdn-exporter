package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = ":9666"
	}
	username := os.Getenv("USERNAME")
	if username == "" {
		log.Fatal("no username provided")
	}
	password := os.Getenv("PASSWORD")
	if password == "" {
		log.Fatal("no password provided")
	}
	timestampMetrics := os.Getenv("TIMESTAMP_METRICS") == "true"

	c := Client{}
	err := c.Login(username, password)
	if err != nil {
		log.Fatalf("error logging in: %v", err)
	}

	coll := &ultraCDNCollector{
		Client:           &c,
		TimestampMetrics: timestampMetrics,
	}

	// Call Login every hour to stay logged in.
	relogin := time.NewTicker(60 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-relogin.C:
				err := c.Login(username, password)
				if err != nil {
					log.Fatalf("error relogging: %v", err)
				}

			case <-quit:
				relogin.Stop()
				return
			}
		}
	}()

	prometheus.MustRegister(coll)

	log.Printf("Listening on port %s\n", port)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(port, nil))
}
