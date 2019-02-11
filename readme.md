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