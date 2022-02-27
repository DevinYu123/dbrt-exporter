package collector

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus"
)

type Mysqlrt struct {
	Mysqlinstences string
	Mysqluser      string
	Mysqlpassword  string
	Mysqldb        string
	Mutithreads    int
}

func (m Mysqlrt) GetMysqlrt(mysqlinstence string) (map[string]float64, error) {
	rt := map[string]float64{
		"connect_rt": -1,
		"delete_rt":  -1,
		"insert_rt":  -1,
		"update_rt":  -1,
		"select_rt":  -1,
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=5s&readTimeout=5s", m.Mysqluser, m.Mysqlpassword, strings.Split(mysqlinstence, "@@")[1], m.Mysqldb)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return rt, errors.New(err.Error())
	}
	defer db.Close()

	rt_begin := time.Now().UnixNano()
	if err = db.Ping(); err != nil {
		return rt, errors.New(err.Error())
	}
	rt_end := time.Now().UnixNano()
	rt["connect_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)

	if strings.Split(mysqlinstence, "@")[0] == "rw" {
		rt_begin = time.Now().UnixNano()
		if _, err := db.Exec("delete from dbrt;"); err != nil {
			return rt, errors.New(err.Error())
		}
		rt_end = time.Now().UnixNano()
		rt["delete_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)

		rt_begin = time.Now().UnixNano()
		if _, err = db.Exec("insert into dbrt value(1,now());"); err != nil {
			return rt, errors.New(err.Error())
		}
		rt_end = time.Now().UnixNano()
		rt["insert_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)

		rt_begin = time.Now().UnixNano()
		if _, err = db.Exec("update dbrt set dbrt_time = now();"); err != nil {
			return rt, errors.New(err.Error())
		}
		rt_end = time.Now().UnixNano()
		rt["update_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)

	} else {
		rt["delete_rt"] = 0
		rt["insert_rt"] = 0
		rt["update_rt"] = 0
	}

	rt_begin = time.Now().UnixNano()
	rs, err := db.Query("select 1 from dbrt;")
	rt_end = time.Now().UnixNano()
	if err != nil {
		return rt, errors.New(err.Error())
	}
	rt["select_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)

	rs.Close()

	return rt, nil
}

func (Mysqlrt) Name() string {
	return "mysql"
}

func (m Mysqlrt) Scrape(ch chan<- prometheus.Metric) error {
	var mywg sync.WaitGroup
	mych := make(chan string, 16)

	for i := 0; i < m.Mutithreads; i++ {
		mywg.Add(1)

		go func() {
			defer mywg.Done()

			for mysqlinstence := range mych {
				myrt, err := m.GetMysqlrt(mysqlinstence)
				if err != nil {
					log.Println(err)
				}

				ch <- prometheus.MustNewConstMetric(
					//这里的label是固定标签 我们可以通过
					NewDesc("mysql_response_time", "connect_us", "mysql connect response time", []string{"instence_type"}, prometheus.Labels{"targetinstance": strings.Split(mysqlinstence, "@@")[1]}),
					prometheus.GaugeValue,
					myrt["connect_rt"],
					//动态标签的值 可以有多个动态标签
					strings.Split(mysqlinstence, "@@")[0],
				)

				ch <- prometheus.MustNewConstMetric(
					NewDesc("mysql_response_time", "select_us", "mysql select response time", []string{"instence_type"}, prometheus.Labels{"targetinstance": strings.Split(mysqlinstence, "@@")[1]}),
					prometheus.GaugeValue,
					myrt["select_rt"],
					strings.Split(mysqlinstence, "@@")[0],
				)

				if strings.Split(mysqlinstence, "@@")[0] == "rw" {
					ch <- prometheus.MustNewConstMetric(
						NewDesc("mysql_response_time", "delete_us", "mysql delete response time", []string{"instence_type"}, prometheus.Labels{"targetinstance": strings.Split(mysqlinstence, "@@")[1]}),
						prometheus.GaugeValue,
						myrt["delete_rt"],
						"rw",
					)

					ch <- prometheus.MustNewConstMetric(
						NewDesc("mysql_response_time", "insert_us", "mysql insert response time", []string{"instence_type"}, prometheus.Labels{"targetinstance": strings.Split(mysqlinstence, "@@")[1]}),
						prometheus.GaugeValue,
						myrt["insert_rt"],
						"rw",
					)

					ch <- prometheus.MustNewConstMetric(
						NewDesc("mysql_response_time", "update_us", "mysql update response time", []string{"instence_type"}, prometheus.Labels{"targetinstance": strings.Split(mysqlinstence, "@@")[1]}),
						prometheus.GaugeValue,
						myrt["update_rt"],
						"rw",
					)
				}
			}
		}()
	}

	for _, mysqlinstence := range strings.Split(m.Mysqlinstences, ",") {
		if len(strings.Split(mysqlinstence, "@@")) == 2 {
			mych <- mysqlinstence
		} else {
			log.Println("The format of MySQL instance is error:", mysqlinstence)
		}
	}

	close(mych)
	mywg.Wait()

	return nil
}
