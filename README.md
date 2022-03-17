# dbrt-exporter
Collect the response time of MySQL, mongodb and redis.

![image](https://github.com/DevinYu123/dbrt-exporter/blob/main/grafana.png)

# Setup
1.Init

Execute the SQL on the target MySQL or MongoDB instance in the init.sql file.

2.Startup

./dbrt_exporter --config=./dbrt.toml

# help
./dbrt_exporter --help
