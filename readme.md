# ultracdn-exporter

A Prometheus exporter for Leaseweb UltraCDN Metrics.

API Specification can be found [here](https://portal.leasewebultracdn.com/apidoc.html).

## Metrics

The exporter exports the following metrics for each distribution group:

| name                | unit | description |
|---------------------|:----:|:-----------:|
|leaseweb_ultracdn_delivered_bytes       |bytes |Total number of bytes delivered in the last 5 minutes.        |
|leaseweb_ultracdn_requests_total        |total |Total number of requests received in the last 5 minutes.      |
|leaseweb_ultracdn_bandwidth_per_second_bytes         |B/s   |Total bandwidth per second summarized over the last 5 minutes.|
|leaseweb_ultracdn_cachehits_per_requests_ratio    |total |Ratio of cachehits per requests in the last 5 minutes.        |
|leaseweb_ultracdn_status_2xx_total |total |Total number of 2xx status codes sent in the last 5 minutes.  |
|leaseweb_ultracdn_status_4xx_total |total |Total number of 4xx status codes sent in the last 5 minutes.  |
|leaseweb_ultracdn_status_5xx_total |total |Total number of 5xx status codes sent in the last 5 minutes.  |

Each metric is exported with `distribution_group` and `distribution_group_id` as labels.

## Configuration

The exporter expects username and password for an UltraCDN account with read permissions to be passed via environment variables. Additonally, a port can be chosen:

| ENV | value | default |
|-----|:-------|:-------:|
|USERNAME | account username | _none_ |
|PASSWORD | account password | _none_ |
|PORT     | port to listen on| 9666   |
|TIMESTAMP_METRICS| add timestamp of datapoint to metrics | "false"|

## Usage

Simply point Prometheus to scrape from `host:port/metrics`.  
It does not make sense to scrape more frequently than 5 minutes, as new metrics will only be available in 5 minute intervals from UltraCDN.  
Note that all metrics will have a lag of ~20 minutes, as metrics are not available earlier from UltraCDN.
Since Prometheus usually does not accept too old metrics, they are not timestamped either with the correct time for the value. 
If you want the original timestamps of datapoints to be exported to Prometheus, set the environment variable `TIMESTAMP_METRICS=true` when running the exporter.
Be warned that, depending on configuration, Prometheus may chose to not ingest the metrics in that case.
