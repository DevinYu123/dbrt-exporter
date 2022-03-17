# dbrt-exporter
Collect the response time of MySQL, mongodb and redis.

![image](https://github.com/DevinYu123/dbrt-exporter/blob/main/grafana.png)

# Startup
./dbrt_exporter --config=./dbrt.toml

# help
./dbrt_exporter --help
Usage of ./dbrt_exporter:
--addr string              The address to listen on for HTTP requests. ENV DBRTADDR. (default ":8306")
--config string            configuration file specifying. ENV DBRTCONFIG.
--enables string           The swich for collectting metrics of mysql, redis or mongodb. separated by commas. ENV DBRTENABLES.
--mongo.db string          The mongo database. ENV DBRTMONGODB. (default "dbrt")
--mongo.instances string   IP:PORT, mongo instances separated by commas. ENV DBRTMONGOINSTANCES. (default "127.0.0.1:27017")
--mongo.password string    The mongo password. ENV DBRTMONGOPASSWORD. (default "dbrt")
--mongo.threads int        Multi-threads for scrape. ENV DBRTMONGOTHREADS. (default 4)
--mongo.user string        The mongo user. ENV DBRTMONGOUSER. (default "dbrt")
--mysql.db string          The mysql database. ENV DBRTMYSQLDB. (default "dbrt")
--mysql.instances string   ro/rw@@IP:PORT, mysql instances separated by commas. ENV DBRTMYSQLINSTANCES. (default "rw@@127.0.0.1:3306")
--mysql.password string    The mysql password. ENV DBRTMYSQLPASSWORD. (default "dbrt")
--mysql.threads int        Multi-threads for scrape. ENV DBRTMYSQLTHREADS. (default 4)
--mysql.user string        The mysql user. ENV DBRTMYSQLUSER. (default "dbrt")
--redis.instances string   PASSWORD@@IP:PORT, redis instances separated by commas. ENV DBRTREDISINSTANCES. (default "password@@127.0.0.1:6379")
--redis.threads int        Multi-threads for scrape. ENV DBRTREDISTHREADS. (default 4)
