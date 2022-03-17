package collector

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongort struct {
	Mongoinstances *string
	Mongouser      *string
	Mongopassword  *string
	Mongodb        *string
	Mutithreads    *int
}

type DbrtFind struct {
	Dbrt int64
}

type IsMaster struct {
	IsMaster bool `bson:"ismaster"`
}

func (m Mongort) GetMongort(mongoinstance string) (map[string]float64, error) {
	rt := map[string]float64{
		"connect_rt": -1,
		"delete_rt":  -1,
		"insert_rt":  -1,
		"update_rt":  -1,
		"find_rt":    -1,
	}

	var (
		dbrtfind DbrtFind
		ismaster IsMaster
	)

	uri := fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s", *m.Mongouser, *m.Mongopassword, mongoinstance, *m.Mongodb, *m.Mongodb)

	opts := options.Client().ApplyURI(uri)
	opts.SetAppName("dbrt-exporter")
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetDirect(true)
	opts.SetMaxPoolSize(1)
	opts.SetMinPoolSize(1)
	opts.SetMaxConnecting(1)

	// 发起链接
	rt_begin := time.Now().UnixNano()
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		return rt, errors.New(err.Error())
	}

	defer client.Disconnect(context.TODO())

	// 判断服务是不是可用
	if err = client.Ping(context.TODO(), nil); err != nil {
		return rt, errors.New(err.Error())
	}
	rt_end := time.Now().UnixNano()
	rt["connect_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)

	collection := client.Database(*m.Mongodb).Collection("dbrt")

	rt_begin = time.Now().UnixNano()
	err = collection.FindOne(context.TODO(), bson.M{}).Decode(&dbrtfind)
	if err != nil {
		return rt, errors.New(err.Error())
	}
	rt_end = time.Now().UnixNano()
	rt["find_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)

	err = client.Database("wdop_service_available").RunCommand(context.TODO(), bson.D{{Key: "isMaster", Value: 1}}).Decode(&ismaster)
	if err != nil {
		return rt, errors.New(err.Error())
	}

	if ismaster.IsMaster {
		rt_begin = time.Now().UnixNano()
		insertResult, err := collection.InsertOne(context.TODO(), DbrtFind{time.Now().Unix()})
		if err != nil {
			return rt, errors.New(err.Error())
		}
		rt_end = time.Now().UnixNano()
		rt["insert_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)

		update := bson.M{"$set": DbrtFind{time.Now().Unix()}}
		rt_begin = time.Now().UnixNano()
		_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": insertResult.InsertedID}, update)
		if err != nil {
			return rt, errors.New(err.Error())
		}
		rt_end = time.Now().UnixNano()
		rt["update_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)

		rt_begin = time.Now().UnixNano()
		_, err = collection.DeleteOne(context.TODO(), bson.M{})
		if err != nil {
			return rt, nil
		}
		rt_end = time.Now().UnixNano()
		rt["delete_rt"] = math.Ceil(float64(rt_end-rt_begin) / 1000)
	} else {
		rt["delete_rt"] = -2
		rt["insert_rt"] = -2
		rt["update_rt"] = -2
	}

	return rt, nil
}

func (Mongort) Name() string {
	return "mongo"
}

func (m Mongort) Scrape(ch chan<- prometheus.Metric) error {
	var mywg sync.WaitGroup
	mych := make(chan string, 16)

	for i := 0; i < *m.Mutithreads; i++ {
		mywg.Add(1)

		go func() {
			defer mywg.Done()

			for mongoinstance := range mych {
				myrt, err := m.GetMongort(mongoinstance)
				if err != nil {
					logrus.Errorf("Scraper mongo %s error: %v", mongoinstance, err)
				}

				ch <- prometheus.MustNewConstMetric(
					NewDesc("mongo_response_time", "connect_us", "mongo connect response time", []string{}, prometheus.Labels{"targetinstance": mongoinstance}),
					prometheus.GaugeValue,
					myrt["connect_rt"],
				)

				ch <- prometheus.MustNewConstMetric(
					NewDesc("mongo_response_time", "find_us", "mongo find response time", []string{}, prometheus.Labels{"targetinstance": mongoinstance}),
					prometheus.GaugeValue,
					myrt["find_rt"],
				)

				if myrt["delete_rt"] != -2 {
					ch <- prometheus.MustNewConstMetric(
						NewDesc("mongo_response_time", "delete_us", "mongo delete response time", []string{}, prometheus.Labels{"targetinstance": mongoinstance}),
						prometheus.GaugeValue,
						myrt["delete_rt"],
					)

					ch <- prometheus.MustNewConstMetric(
						NewDesc("mongo_response_time", "insert_us", "mongo insert response time", []string{}, prometheus.Labels{"targetinstance": mongoinstance}),
						prometheus.GaugeValue,
						myrt["insert_rt"],
					)

					ch <- prometheus.MustNewConstMetric(
						NewDesc("mongo_response_time", "update_us", "mongo update response time", []string{}, prometheus.Labels{"targetinstance": mongoinstance}),
						prometheus.GaugeValue,
						myrt["update_rt"],
					)
				}
			}
		}()
	}

	for _, mongoinstance := range strings.Split(*m.Mongoinstances, ",") {
		mych <- mongoinstance
	}

	close(mych)
	mywg.Wait()

	return nil
}
