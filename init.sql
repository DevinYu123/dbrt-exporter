#MongoDB
use dbrt;
db.dbrt.insert({'dbrt': NumberLong(1)});
db.createUser({user: 'dbrt', pwd: 'dbrt', roles: [{'role' : 'readWrite', 'db' : 'dbrt'},]});

#MySQL
create database dbrt;
create table dbrt.dbrt (id int(11) NOT NULL DEFAULT '0',dbrt_time datetime DEFAULT NULL,PRIMARY KEY (id));
insert into dbrt.dbrt value(1,now());
create user dbrt@'%' identified by 'dbrt';
grant select,insert,update,delete on dbrt.* to dbrt@'%';
