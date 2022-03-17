package collector

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	name      string = "dbrt_exporter"
	namespace string = "dbrt"
	exporter  string = "exporter"
)

func Name() string {
	return name
}

type Exporter struct {
	metrics  Metrics
	scrapers []Scraper
}

type Metrics struct {
	//收集指标的总次数
	TotalScrapes prometheus.Counter
	//收集指标中发生错误的次数
	ScrapeErrors *prometheus.CounterVec
	//最后一次是否发生了错误
	Error prometheus.Gauge
}

//fqname会把namespace subsystem name拼接起来
//传入动态以及静态标签 设置标签
func NewDesc(subsystem, name, help string, movinglabel []string, label prometheus.Labels) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, name),
		help, movinglabel, label,
	)
}

//判断*exporter是否实现了collector这个接口的所有方法
var _ prometheus.Collector = (*Exporter)(nil)

func New(metrics Metrics, scrapers []Scraper) *Exporter {
	return &Exporter{
		metrics:  metrics,
		scrapers: scrapers,
	}
}

func NewMetrics() Metrics {
	subsystem := exporter
	return Metrics{
		TotalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "scrapes_total",
			Help:      "Total number of times  was scraped for metrics.",
		}),
		ScrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occurred scraping .",
		}, []string{"collector"}),
		Error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics  resulted in an error (1 for error, 0 for success).",
		}),
	}
}

//添加描述符
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.metrics.TotalScrapes.Desc()
	ch <- e.metrics.Error.Desc()
	e.metrics.ScrapeErrors.Describe(ch)
}

//收集指标
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.scrape(ch)
	ch <- e.metrics.TotalScrapes
	ch <- e.metrics.Error
	e.metrics.ScrapeErrors.Collect(ch)
}

//通过例程并发的收集指标 需要加waitgroup
func (e *Exporter) scrape(ch chan<- prometheus.Metric) {
	var (
		wg  sync.WaitGroup
		err error
	)

	defer wg.Wait()
	for _, scraper := range e.scrapers {
		wg.Add(1)
		//使用匿名函数 并且并发的收集指标
		go func(scraper Scraper) {
			defer wg.Done()
			label := scraper.Name()
			err = scraper.Scrape(ch)
			if err != nil {
				logrus.Errorf("scraper %s: %s", scraper.Name(), err)
				e.metrics.ScrapeErrors.WithLabelValues(label).Inc()
				e.metrics.Error.Set(1)
			}
		}(scraper)
	}
}
