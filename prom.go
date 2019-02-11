package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"time"
)

var (
	bytesdeliveredDesc = prometheus.NewDesc(
		"leaseweb_ultracdn_delivered_bytes",
		"Total number of bytes delivered in the last 5 minutes.",
		[]string{"distribution_group", "distribution_group_id"}, nil)
	requestscountDesc = prometheus.NewDesc(
		"leaseweb_ultracdn_requests_total",
		"Total number of requests received in the last 5 minutes.",
		[]string{"distribution_group", "distribution_group_id"}, nil)
	bandwidthbpsDesc = prometheus.NewDesc(
		"leaseweb_ultracdn_bandwidth_per_second_bytes",
		"Total bandwidth per second summarized over the last 5 minutes.",
		[]string{"distribution_group", "distribution_group_id"}, nil)
	cachehit_requestsDesc = prometheus.NewDesc(
		"leaseweb_ultracdn_cachehits_per_requests_ratio",
		"Ratio of cachehits per requests in the last 5 minutes.",
		[]string{"distribution_group", "distribution_group_id"}, nil)
	statuscode_2xx_countDesc = prometheus.NewDesc(
		"leaseweb_ultracdn_status_2xx_total",
		"Total number of 2xx status codes sent in the last 5 minutes.",
		[]string{"distribution_group", "distribution_group_id"}, nil)
	statuscode_4xx_countDesc = prometheus.NewDesc(
		"leaseweb_ultracdn_status_4xx_total",
		"Total number of 4xx status codes sent in the last 5 minutes.",
		[]string{"distribution_group", "distribution_group_id"}, nil)
	statuscode_5xx_countDesc = prometheus.NewDesc(
		"leaseweb_ultracdn_status_5xx_total",
		"Total number of 5xx status codes sent in the last 5 minutes.",
		[]string{"distribution_group", "distribution_group_id"}, nil)
)

var descs = map[string]*prometheus.Desc{
	"bytesdelivered":       bytesdeliveredDesc,
	"requestscount":        requestscountDesc,
	"bandwidthbps":         bandwidthbpsDesc,
	"cachehit_requests":    cachehit_requestsDesc,
	"statuscode_2xx_count": statuscode_2xx_countDesc,
	"statuscode_4xx_count": statuscode_4xx_countDesc,
	"statuscode_5xx_count": statuscode_5xx_countDesc,
}

type ultraCDNCollector struct {
	Client *Client
}

func (c *ultraCDNCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *ultraCDNCollector) Collect(ch chan<- prometheus.Metric) {
	for _, dg := range c.Client.DistGroups {
		for t, desc := range descs {
			m, err := c.Client.FetchMetric(dg.ID, t)

			if err != nil {
				log.Printf("error fetching Metric %s for distributiongroup %s: %v", t, dg.ID, err)
				break
			}

			if len(m.Points) == 0 {
				break
			}

			p := m.Points[0]
			ch <- prometheus.NewMetricWithTimestamp(time.Unix(int64(p.Timestamp), 0),
				prometheus.MustNewConstMetric(
					desc,
					prometheus.GaugeValue,
					p.Value,
					dg.Name, dg.ID))
		}
	}
}
