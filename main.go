package main

import (
	"dbrt_exporter/collector"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsPath string = "https://github.com/DevinYu123/dbrt-exporter"
	version     string = "v1.0"
	help        bool

	listenaddr     = flag.String("addr", ":8306", "The address to listen on for HTTP requests.")
	enables        = flag.String("enables", "", "The swich for collectting metrics of mysql, redis or mongodb. separated by commas.")
	mysqlinstences = flag.String("mysql-instences", "", "ro/rw@@IP:PORT, mysql instances separated by commas.")
	mysqluser      = flag.String("mysql-user", "dbrt", "The mysql user.")
	mysqlpassword  = flag.String("mysql-password", "dbrt", "The mysql password.")
	mysqldb        = flag.String("mysql-db", "dbrt", "The mysql database.")
	mongoinstences = flag.String("mongo-instences", "", "IP:PORT, mongo instances separated by commas.")
	mongouser      = flag.String("mongo-user", "dbrt", "The mongo user.")
	mongopassword  = flag.String("mongo-password", "dbrt", "The mongo password.")
	mongodb        = flag.String("mongo-db", "dbrt", "The mongo database.")
	redisinstences = flag.String("redis-instences", "", "PASSWORD@@IP:PORT, redis instances separated by commas.")
	mutithreads    = flag.Int("muti-threads", 4, "Multi-threads for scrape.")
)

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	Scrapers := map[collector.Scraper]bool{
		collector.Mysqlrt{
			Mysqlinstences: *mysqlinstences,
			Mysqluser:      *mysqluser,
			Mysqlpassword:  *mysqlpassword,
			Mysqldb:        *mysqldb,
			Mutithreads:    *mutithreads,
		}: false,
		collector.Redisrt{
			Redisinstences: *redisinstences,
			Mutithreads:    *mutithreads,
		}: false,
		collector.Mongort{
			Mongoinstences: *mongoinstences,
			Mongouser:      *mongouser,
			Mongopassword:  *mongopassword,
			Mongodb:        *mongodb,
			Mutithreads:    *mutithreads,
		}: false,
	}

	for scraper := range Scrapers {
		for _, v := range strings.Split(*enables, ",") {
			if v == scraper.Name() {
				Scrapers[scraper] = true
				break
			}
		}
	}

	enabledScrapers := []collector.Scraper{}
	for scraper, enabled := range Scrapers {
		if enabled {
			log.Printf("Scraper enabled %s", scraper.Name())
			enabledScrapers = append(enabledScrapers, scraper)
		}
	}

	exporter := collector.New(collector.NewMetrics(), enabledScrapers)
	prometheus.MustRegister(exporter)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
<head><title>` + collector.Name() + `</title></head>
<body>
<h1><a style="text-decoration:none" href=''>` + collector.Name() + `</a></h1>
<p><a href='` + metricsPath + `'>/metrics</a></p>
<h2>Build</h2>
<pre>` + version + `</pre>
</body>
</html>`))
	})

	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(*listenaddr, nil); err != nil {
		log.Printf("Error occur when start server %v", err)
	}
}
