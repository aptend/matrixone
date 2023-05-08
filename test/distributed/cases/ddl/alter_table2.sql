create database if not exists db;
use db;
drop table if exists t;
create table t (a int, b int, primary key (a));

insert into t values (1, 1), (2, 2), (3, 3), (4,4), (5, 5);

-- @separator:table
select mo_ctl('dn', 'inspect', 'addc -d db -b t -n aplus -t samllint -p 2');

show columns from t;

select * from t;

insert into t values (6, 6,6), (7,7,7);

-- @separator:table
select mo_ctl('dn', 'flush', 'db.t');

select aplus from t;

update t set aplus = 11 where a = 1;
update t set aplus = 61 where a = 6;

select * from t;


-- @separator:table
select mo_ctl('dn', 'inspect', 'dropc -d db -b t -p 2');

-- @separator:table
select mo_ctl('dn', 'inspect', 'addc -d db -b t -n aminus -t samllint -p 0');

show columns from t;

select attname, attr_seqnum from mo_catalog.mo_columns where att_relname = "t";
select relname, rel_version from mo_catalog.mo_tables where reldatabase = "db";

insert into t values (12, 12, 12);

select * from t;

show tables;
-- @separator:table
select mo_ctl('dn', 'inspect', 'renamet -d db -o t -n newt');

select * from t;
select * from newt;

select attname, attr_seqnum from mo_catalog.mo_columns where att_relname = "t";
select attname, attr_seqnum from mo_catalog.mo_columns where att_relname = "newt";
select relname, rel_version from mo_catalog.mo_tables where reldatabase = "db";

-- check zonmeap filter
select a,b from newt where a > 4 and b < 10;

drop database db;
