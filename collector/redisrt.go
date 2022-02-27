package collector

import (
	"errors"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
)

type Redisrt struct {
	Redisinstences string
	Mutithreads    int
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
		log.Println("dbrt_exporter does not exists")
	} else if err != nil {
		return rt, errors.New(err.Error())
	}
	rt["get_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)

	rt_begin = time.Now().UnixNano()
	err = client.Set("dbrt_exporter", "dbrt_exporter", 0).Err()
	if err != nil {
		return rt, nil
	}
	rt_end = time.Now().UnixNano()
	rt["set_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)

	return rt, nil
}

func (Redisrt) Name() string {
	return "redis"
}

func (r Redisrt) Scrape(ch chan<- prometheus.Metric) error {
	var mywg sync.WaitGroup
	mych := make(chan string, 16)

	for i := 0; i < r.Mutithreads; i++ {
		mywg.Add(1)

		go func() {
			defer mywg.Done()

			for redisinstence := range mych {
				myrt, err := r.GetRedisrt(redisinstence)
				if err != nil {
					log.Println(err)
				}

				ch <- prometheus.MustNewConstMetric(
					NewDesc("redis_response_time", "connect_us", "redis connect response time", []string{}, prometheus.Labels{"targetinstance": strings.Split(redisinstence, "@@")[1]}),
					prometheus.GaugeValue,
					myrt["connect_rt"],
				)

				ch <- prometheus.MustNewConstMetric(
					NewDesc("redis_response_time", "get_us", "redis get response time", []string{}, prometheus.Labels{"targetinstance": strings.Split(redisinstence, "@@")[1]}),
					prometheus.GaugeValue,
					myrt["get_rt"],
				)

				ch <- prometheus.MustNewConstMetric(
					NewDesc("redis_response_time", "set_us", "redis set response time", []string{}, prometheus.Labels{"targetinstance": strings.Split(redisinstence, "@@")[1]}),
					prometheus.GaugeValue,
					myrt["set_rt"],
				)
			}
		}()
	}

	for _, redisinstence := range strings.Split(r.Redisinstences, ",") {
		if len(strings.Split(redisinstence, "@@")) == 2 {
			mych <- redisinstence
		} else {
			log.Println("The format of Redis instance is error:", redisinstence)
		}
	}

	close(mych)
	mywg.Wait()

	return nil
}
