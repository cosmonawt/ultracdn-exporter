package main

import "github.com/prometheus/client_golang/prometheus"

const (
	promNamespace = "leaseweb"
	promSubsystem = "ultracdn"
)

type ultraCDNCollector struct {
	Desc   *prometheus.Desc
	Metric prometheus.Metric
}

func (c *ultraCDNCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Desc
}

func (c *ultraCDNCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- c.Metric
}

//desc := prometheus.NewDesc(
////"temperature_kelvin",
////"Current temperature in Kelvin.",
////nil, nil,
////)
////
////temperatureReportedByExternalSystem := 298.15
////timeReportedByExternalSystem := time.Date(2009, time.November, 10, 23, 0, 0, 12345678, time.UTC)
////m := prometheus.NewMetricWithTimestamp(
////timeReportedByExternalSystem,
////prometheus.MustNewConstMetric(
////desc, prometheus.GaugeValue, temperatureReportedByExternalSystem,
////),
////)
////
////col := &ultraCDNCollector{desc, m}
////
////prometheus.MustRegister(col)
