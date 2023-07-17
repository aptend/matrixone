drop database if exists data;
create database if not exists data;
use data;
drop table if exists sim_track_main_local;
CREATE TABLE if not exists sim_track_main_local
(
    `vehicle_id` int unsigned COMMENT '车辆id',
    `veh_class` TINYINT COMMENT '车辆类型（1：小汽车，2：公交车）',
    unique key (vehicle_id, veh_class),
);

insert into sim_track_main_local values (12770,6,"1552-06-21 08:36:13","1154-01-26 02:0:37","1653-11-16 09:53:14",1,11499.5349199356,25602,28112,15209,"252.22003532392554",183.4477220678895,6986,5,13,1877,"6725","1050","1079.9010193364684","1595-10-16 15:37:20","2014-06-20 14:33:23","opqrstuvwx","cdefghigklmnopqrst");

insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
insert into sim_track_main_local select * from sim_track_main_local;
