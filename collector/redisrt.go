package collector

import (
	"errors"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type Redisrt struct {
	Redisinstances *string
	Mutithreads    *int
}

func (Redisrt) GetRedisrt(redisinstance string) (map[string]float64, error) {
	rt := map[string]float64{
		"connect_rt": -1,
		"get_rt":     -1,
		"set_rt":     -1,
	}

	client := redis.NewClient(&redis.Options{
		Addr:     strings.Split(redisinstance, "@@")[1],
		Password: strings.Split(redisinstance, "@@")[0],
		DB:       0,
	})

	defer client.Close()

	rt_begin := time.Now().UnixNano()
	if err := client.Ping().Err(); err != nil {
		return rt, errors.New(err.Error())
	}
	rt_end := time.Now().UnixNano()
	rt["connect_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)

	rt_begin = time.Now().UnixNano()
	_, err := client.Get("dbrt_exporter").Result()
	rt_end = time.Now().UnixNano()
	if err == redis.Nil {
	} else if err != nil {
		return rt, errors.New(err.Error())
	}
	rt["get_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)

	if strings.Split(client.Info("Replication").String(), "\r\n")[1] == "role:master" {
		rt_begin = time.Now().UnixNano()
		err = client.Set("dbrt_exporter", time.Now().Unix(), 0).Err()
		if err != nil {
			return rt, errors.New(err.Error())
		}
		rt_end = time.Now().UnixNano()
		rt["set_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)
	} else {
		rt["set_rt"] = -2
	}

	return rt, nil
}

func (Redisrt) Name() string {
	return "redis"
}

func (r Redisrt) Scrape(ch chan<- prometheus.Metric) error {
	var mywg sync.WaitGroup
	mych := make(chan string, 16)

	for i := 0; i < *r.Mutithreads; i++ {
		mywg.Add(1)

		go func() {
			defer mywg.Done()

			for redisinstance := range mych {
				myrt, err := r.GetRedisrt(redisinstance)
				if err != nil {
					logrus.Errorf("Scraper redis %s error: %v", strings.Split(redisinstance, "@@")[1], err)
				}

				ch <- prometheus.MustNewConstMetric(
					NewDesc("redis_response_time", "connect_us", "redis connect response time", []string{}, prometheus.Labels{"targetinstance": strings.Split(redisinstance, "@@")[1]}),
					prometheus.GaugeValue,
					myrt["connect_rt"],
				)

				ch <- prometheus.MustNewConstMetric(
					NewDesc("redis_response_time", "get_us", "redis get response time", []string{}, prometheus.Labels{"targetinstance": strings.Split(redisinstance, "@@")[1]}),
					prometheus.GaugeValue,
					myrt["get_rt"],
				)

				if myrt["set_rt"] != -2 {
					ch <- prometheus.MustNewConstMetric(
						NewDesc("redis_response_time", "set_us", "redis set response time", []string{}, prometheus.Labels{"targetinstance": strings.Split(redisinstance, "@@")[1]}),
						prometheus.GaugeValue,
						myrt["set_rt"],
					)
				}
			}
		}()
	}

	for _, redisinstance := range strings.Split(*r.Redisinstances, ",") {
		if len(strings.Split(redisinstance, "@@")) == 2 {
			mych <- redisinstance
		} else {
			logrus.Error("The format of Redis instance is error: ", redisinstance)
		}
	}

	close(mych)
	mywg.Wait()

	return nil
}
