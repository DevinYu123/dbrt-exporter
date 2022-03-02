#MongoDB
use xxx
db.createUser({
user: "dbrt",
pwd: "dbrt",
roles: [
{"role" : "readWrite", "db" : "xxx"},
]
})
db.dbrt.insert({})

#MySQL
create user dbrt@'172.21.%' identified by 'dbrt';
grant select,insert,update,delete on xxx.* to dbrt@'172.21.%';
create database xxx;
CREATE TABLE xxx.dbrt (
  `id` int(11) NOT NULL DEFAULT '0',
  `dbrt_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
);
