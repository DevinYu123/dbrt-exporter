package main

import (
	"dbrt_exporter/collector"
	"net/http"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	metricsPath    string = "https://github.com/DevinYu123/dbrt-exporter"
	version        string = "v1.1"
	listenaddr     string
	enables        string
	mysqlInstances string
	mysqlUser      string
	mysqlPassword  string
	mysqlDB        string
	mysqlThreads   int
	mongoInstances string
	mongoUser      string
	mongoPassword  string
	mongoDB        string
	mongoThreads   int
	redisinstances string
	redisThreads   int
)

func init() {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02T15:04:05"
	logrus.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	pflag.String("addr", ":8306", "The address to listen on for HTTP requests. ENV DBRTADDR.")
	pflag.String("config", "", "configuration file specifying. ENV DBRTCONFIG.")
	pflag.String("enables", "", "The swich for collectting metrics of mysql, redis or mongodb. separated by commas. ENV DBRTENABLES.")
	pflag.String("mysql.instances", "rw@@127.0.0.1:3306", "ro/rw@@IP:PORT, mysql instances separated by commas. ENV DBRTMYSQLINSTANCES.")
	pflag.String("mysql.user", "dbrt", "The mysql user. ENV DBRTMYSQLUSER.")
	pflag.String("mysql.password", "dbrt", "The mysql password. ENV DBRTMYSQLPASSWORD.")
	pflag.String("mysql.db", "dbrt", "The mysql database. ENV DBRTMYSQLDB.")
	pflag.Int("mysql.threads", 4, "Multi-threads for scrape. ENV DBRTMYSQLTHREADS.")
	pflag.String("mongo.instances", "127.0.0.1:27017", "IP:PORT, mongo instances separated by commas. ENV DBRTMONGOINSTANCES.")
	pflag.String("mongo.user", "dbrt", "The mongo user. ENV DBRTMONGOUSER.")
	pflag.String("mongo.password", "dbrt", "The mongo password. ENV DBRTMONGOPASSWORD.")
	pflag.String("mongo.db", "dbrt", "The mongo database. ENV DBRTMONGODB.")
	pflag.Int("mongo.threads", 4, "Multi-threads for scrape. ENV DBRTMONGOTHREADS.")
	pflag.String("redis.instances", "password@@127.0.0.1:6379", "PASSWORD@@IP:PORT, redis instances separated by commas. ENV DBRTREDISINSTANCES.")
	pflag.Int("redis.threads", 4, "Multi-threads for scrape. ENV DBRTREDISTHREADS.")

	viper.BindPFlags(pflag.CommandLine)
	pflag.Parse()

	viper.BindEnv("addr", "DBRTADDR")
	viper.BindEnv("config", "DBRTCONFIG")
	viper.BindEnv("enables", "DBRTENABLES")
	viper.BindEnv("mysql.instances", "DBRTMYSQLINSTANCES")
	viper.BindEnv("mysql.user", "DBRTMYSQLUSER")
	viper.BindEnv("mysql.password", "DBRTMYSQLPASSWORD")
	viper.BindEnv("mysql.db", "DBRTMYSQLDB")
	viper.BindEnv("mysql.threads", "DBRTMYSQLTHREADS")
	viper.BindEnv("mongo.instances", "DBRTMONGOINSTANCES")
	viper.BindEnv("mongo.user", "DBRTMONGOUSER")
	viper.BindEnv("mongo.password", "DBRTMONGOPASSWORD")
	viper.BindEnv("mongo.db", "DBRTMONGODB")
	viper.BindEnv("mongo.threads", "DBRTMONGOTHREADS")
	viper.BindEnv("redis.instances", "DBRTREDISINSTANCES")
	viper.BindEnv("redis.threads", "DBRTREDISTHREADS")
}

func readConfig() {
	listenaddr = viper.GetString("addr")
	enables = viper.GetString("enables")
	mysqlInstances = viper.GetString("mysql.instances")
	mysqlUser = viper.GetString("mysql.user")
	mysqlPassword = viper.GetString("mysql.password")
	mysqlDB = viper.GetString("mysql.db")
	mysqlThreads = viper.GetInt("mysql.threads")
	mongoInstances = viper.GetString("mongo.instances")
	mongoUser = viper.GetString("mongo.user")
	mongoPassword = viper.GetString("mongo.password")
	mongoDB = viper.GetString("mongo.db")
	mongoThreads = viper.GetInt("mongo.threads")
	redisinstances = viper.GetString("redis.instances")
	redisThreads = viper.GetInt("redis.threads")
}

func main() {
	if viper.GetString("config") != "" {
		viper.SetConfigFile(viper.GetString("config"))
		err := viper.ReadInConfig()
		if err != nil {
			logrus.Fatal("Read config file failed: ", err)
		}

		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			readConfig()
		})
	}

	readConfig()

	enabledScrapers := []collector.Scraper{}
	for _, v := range strings.Split(enables, ",") {
		if v == "mysql" {
			enabledScrapers = append(enabledScrapers, collector.Mysqlrt{
				Mysqlinstances: &mysqlInstances,
				Mysqluser:      &mysqlUser,
				Mysqlpassword:  &mysqlPassword,
				Mysqldb:        &mysqlDB,
				Mutithreads:    &mysqlThreads,
			})
			logrus.Info("Scraper enabled mysql. You need to execute the following commands to create table and user for MySQL.")
			logrus.Infof("create database %s;", mysqlDB)
			logrus.Infof("create table %s.dbrt (id int(11) NOT NULL DEFAULT '0',dbrt_time datetime DEFAULT NULL,PRIMARY KEY (id));", mysqlDB)
			logrus.Infof("insert into %s.dbrt value(1,now());", mysqlDB)
			logrus.Infof("create user %s@'%%' identified by 'xxx';", mysqlUser)
			logrus.Infof("grant select,insert,update,delete on %s.dbrt to %s@'%%';", mysqlDB, mysqlUser)
		} else if v == "mongo" {
			enabledScrapers = append(enabledScrapers, collector.Mongort{
				Mongoinstances: &mongoInstances,
				Mongouser:      &mongoUser,
				Mongopassword:  &mongoPassword,
				Mongodb:        &mongoDB,
				Mutithreads:    &mongoThreads,
			})
			logrus.Info("Scraper enabled mongo. You need to execute the following commands to create table and user for MongoDB.")
			logrus.Infof("use %s;", mongoDB)
			logrus.Infof("db.dbrt.insert({'dbrt': NumberLong(1)});")
			logrus.Infof("db.createUser({user: '%s', pwd: 'xxx', roles: [{'role' : 'readWrite', 'db' : '%s'},]});", mongoUser, mongoDB)
		} else if v == "redis" {
			enabledScrapers = append(enabledScrapers, collector.Redisrt{
				Redisinstances: &redisinstances,
				Mutithreads:    &redisThreads,
			})
			logrus.Info("Scraper enabled redis.")
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
	if err := http.ListenAndServe(listenaddr, nil); err != nil {
		logrus.Error("Error occur when start server: ", err)
	}
}
