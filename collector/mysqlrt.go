package collector

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type Mysqlrt struct {
	Mysqlinstances *string
	Mysqluser      *string
	Mysqlpassword  *string
	Mysqldb        *string
	Mutithreads    *int
}

func (m Mysqlrt) GetMysqlrt(mysqlinstance string) (map[string]float64, error) {
	rt := map[string]float64{
		"connect_rt": -1,
		"delete_rt":  -1,
		"insert_rt":  -1,
		"update_rt":  -1,
		"select_rt":  -1,
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=5s&readTimeout=5s", *m.Mysqluser, *m.Mysqlpassword, strings.Split(mysqlinstance, "@@")[1], *m.Mysqldb)
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

	if strings.Split(mysqlinstance, "@")[0] == "rw" {
		rt_begin = time.Now().UnixNano()
		if _, err := db.Exec("delete from dbrt where id = 2;"); err != nil {
			return rt, errors.New(err.Error())
		}
		rt_end = time.Now().UnixNano()
		rt["delete_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)

		rt_begin = time.Now().UnixNano()
		if _, err = db.Exec("insert into dbrt value(2,now());"); err != nil {
			return rt, errors.New(err.Error())
		}
		rt_end = time.Now().UnixNano()
		rt["insert_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)

		rt_begin = time.Now().UnixNano()
		if _, err = db.Exec("update dbrt set dbrt_time = now() where id = 2;"); err != nil {
			return rt, errors.New(err.Error())
		}
		rt_end = time.Now().UnixNano()
		rt["update_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)

	} else {
		rt["delete_rt"] = -2
		rt["insert_rt"] = -2
		rt["update_rt"] = -2
	}

	rt_begin = time.Now().UnixNano()
	rs, err := db.Query("select * from dbrt where id = 1;")
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

	for i := 0; i < *m.Mutithreads; i++ {
		mywg.Add(1)

		go func() {
			defer mywg.Done()

			for mysqlinstance := range mych {
				myrt, err := m.GetMysqlrt(mysqlinstance)
				if err != nil {
					logrus.Errorf("Scraper mysql %s error: %v", mysqlinstance, err)
				}

				ch <- prometheus.MustNewConstMetric(
					//这里的label是固定标签 我们可以通过
					NewDesc("mysql_response_time", "connect_us", "mysql connect response time", []string{"instance_type"}, prometheus.Labels{"targetinstance": strings.Split(mysqlinstance, "@@")[1]}),
					prometheus.GaugeValue,
					myrt["connect_rt"],
					//动态标签的值 可以有多个动态标签
					strings.Split(mysqlinstance, "@@")[0],
				)

				ch <- prometheus.MustNewConstMetric(
					NewDesc("mysql_response_time", "select_us", "mysql select response time", []string{"instance_type"}, prometheus.Labels{"targetinstance": strings.Split(mysqlinstance, "@@")[1]}),
					prometheus.GaugeValue,
					myrt["select_rt"],
					strings.Split(mysqlinstance, "@@")[0],
				)

				if strings.Split(mysqlinstance, "@@")[0] == "rw" {
					ch <- prometheus.MustNewConstMetric(
						NewDesc("mysql_response_time", "delete_us", "mysql delete response time", []string{"instance_type"}, prometheus.Labels{"targetinstance": strings.Split(mysqlinstance, "@@")[1]}),
						prometheus.GaugeValue,
						myrt["delete_rt"],
						"rw",
					)

					ch <- prometheus.MustNewConstMetric(
						NewDesc("mysql_response_time", "insert_us", "mysql insert response time", []string{"instance_type"}, prometheus.Labels{"targetinstance": strings.Split(mysqlinstance, "@@")[1]}),
						prometheus.GaugeValue,
						myrt["insert_rt"],
						"rw",
					)

					ch <- prometheus.MustNewConstMetric(
						NewDesc("mysql_response_time", "update_us", "mysql update response time", []string{"instance_type"}, prometheus.Labels{"targetinstance": strings.Split(mysqlinstance, "@@")[1]}),
						prometheus.GaugeValue,
						myrt["update_rt"],
						"rw",
					)
				}
			}
		}()
	}

	for _, mysqlinstance := range strings.Split(*m.Mysqlinstances, ",") {
		if len(strings.Split(mysqlinstance, "@@")) == 2 {
			mych <- mysqlinstance
		} else {
			logrus.Error("The format of MySQL instance is error: ", mysqlinstance)
		}
	}

	close(mych)
	mywg.Wait()

	return nil
}
