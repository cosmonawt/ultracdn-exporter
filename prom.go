package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"sync"
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

type metricCache struct {
	c map[DistributionGroup]map[string]Metric
	sync.RWMutex
}

var cache = metricCache{
	c: make(map[DistributionGroup]map[string]Metric),
}

type ultraCDNCollector struct {
	Client           *Client
	TimestampMetrics bool
}

func (c *ultraCDNCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *ultraCDNCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	for _, distGroup := range c.Client.DistGroups {
		cache.Lock()
		if cache.c[distGroup] == nil {
			cache.c[distGroup] = map[string]Metric{}
		}
		cache.Unlock()

		for target, desc := range descs {
			wg.Add(1)
			go func(distGroup DistributionGroup, target string, desc *prometheus.Desc) {
				metric, err := c.Client.FetchMetric(distGroup.ID, target)
				if err != nil {
					log.Printf("error fetching Metric %s for distributiongroup %s: %v", target, distGroup.ID, err)
				}

				// If we can'target scrape metrics, we use the ones from cache to avoid a discontinued metric.
				// If cache is empty, we use a 0 metric for the same reason.
				// We multiply the local timestamp by 1000 because leaseweb's timestamps are already in milliseconds
				// Otherwise Prometheus will discard them if the timestamps differ in length
				if len(metric.Points) == 0 {
					cache.RLock()
					pp := cache.c[distGroup][target].Points
					cache.RUnlock()
					if len(pp) == 0 {
						pp = []Point{{
							Value:     float64(0.0),
							Timestamp: int(time.Now().Unix()) * 1000,
						}}
					}
					metric.Points = pp
				}

				// Cache latest entry
				cache.Lock()
				cache.c[distGroup][target] = metric
				cache.Unlock()

				p := metric.Points[0]
				m := prometheus.MustNewConstMetric(
					desc,
					prometheus.GaugeValue,
					p.Value,
					distGroup.Name, distGroup.ID)

				if c.TimestampMetrics {
					m = prometheus.NewMetricWithTimestamp(time.Unix(int64(p.Timestamp)/1000, 0), m)
				}

				ch <- m
				wg.Done()
			}(distGroup, target, desc)
		}
	}
	wg.Wait()
}
