drop database if exists db;
create database db;
use db;
drop table if exists s3t;
create table s3t (a int, b int, c int, primary key(a, b));

insert into s3t select result, 2, 12 from generate_series(1, 30000, 1) g;

alter table s3t add column d int after b;

insert into s3t values (300001, 34, 23, 1);

select count(*) from s3t;

select * from s3t where d = 23;

alter table s3t drop column c;

insert into s3t select result, 2, 12 from generate_series(30002, 60000, 1) g;

select count(d) from s3t;
