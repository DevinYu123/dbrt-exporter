package collector

import "github.com/prometheus/client_golang/prometheus"

type Scraper interface {
	Name() string

	Scrape(ch chan<- prometheus.Metric) error
}
