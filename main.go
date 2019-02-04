package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

const (
	promNamespace = "leaseweb"
	promSubsystem = "ultracdn"
)

var (
	bytesDelivered = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "delivered_bytes",
		Help:      "Total number of bytes delivered in the last 5 minutes.",
	}, []string{"distribution_id", "distribution_name"})
	requestsCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "requests_total",
		Help:      "Total number of requests received in the last 5 minutes.",
	}, []string{"distribution_id", "distribution_name"})
	bandwidthbps = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "bandwidth_per_second_bytes",
		Help:      "Total bandwidth per second summarized over the last 5 minutes.",
	}, []string{"distribution_id", "distribution_name"})
	cachehitRequests = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "cachehits_per_requests_ratio",
		Help:      "Ratio of cachehits per requests in the last 5 minutes.",
	}, []string{"distribution_id", "distribution_name"})
	statusCode2xxCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "status_2xx_total",
		Help:      "Total number of 2xx status codes sent in the last 5 minutes.",
	}, []string{"distribution_id", "distribution_name"})
	statusCode4xxCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "status_4xx_total",
		Help:      "Total number of 4xx status codes sent in the last 5 minutes.",
	}, []string{"distribution_id", "distribution_name"})
	statusCode5xxCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "status_5xx_total",
		Help:      "Total number of 5xx status codes sent in the last 5 minutes.",
	}, []string{"distribution_id", "distribution_name"})
)

func init() {
	prometheus.MustRegister(bytesDelivered)
	prometheus.MustRegister(requestsCount)
	prometheus.MustRegister(bandwidthbps)
	prometheus.MustRegister(cachehitRequests)
	prometheus.MustRegister(statusCode2xxCount)
	prometheus.MustRegister(statusCode4xxCount)
	prometheus.MustRegister(statusCode5xxCount)
}

func main() {
	metrics := map[string]*prometheus.GaugeVec{
		"bytesdelivered":       bytesDelivered,
		"requestscount":        requestsCount,
		"bandwidthbps":         bandwidthbps,
		"cachehit_requests":    cachehitRequests,
		"statuscode_2xx_count": statusCode2xxCount,
		"statuscode_4xx_count": statusCode4xxCount,
		"statuscode_5xx_count": statusCode5xxCount,
	}

	c := client{}
	err := c.login("", "")
	if err != nil {
		log.Fatalf("error logging in: %v", err)
	}

	cid, err := c.getCustomerID()
	if err != nil {
		log.Fatalf("error getting customerID: %v", err)
	}

	dd, err := c.getDistributionGroups(cid)
	if err != nil {
		log.Fatalf("error getting distribution groups: %v", err)
	}

	for _, d := range dd {
		mm, err := c.gatherMetrics(d.ID)
		if err != nil {
			log.Fatalf("error getting metrics for distribution group %s: %v", d.ID, err)
		}
		for _, m := range mm {
			metrics[m.Target].WithLabelValues(d.ID, d.Name).Set(m.Points[1].Value)
		}
	}

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
