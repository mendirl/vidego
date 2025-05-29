drop schema if exists videogo cascade ;

create schema if not exists videogo;

drop table if exists videogo.video;

create table videogo.video
(
    id          serial primary key,
    name        text,
    path        text,
    size        numeric,
    duration    numeric,
    created_at  timestamp,
    updated_at  timestamp,
    deleted_at  timestamp,
    deduplicate boolean default false,
    complete    boolean default false
);

update videogo.video set deduplicate = false where path like '%dedup';

truncate table videogo.video;

select count(1) from videogo.video;


select count(1)
from videogo.video
where deduplicate is false;



select count(1)
from videogo.video
where path like '%dedup';


select *
from videogo.video
where duration in (select duration from videogo.video group by duration having count(1) > 1)
order by duration desc;

select *
from videogo.video
where path not like '%dedup%'
  and duration in (select duration from videogo.video where path not like '%dedup' group by duration having count(1) > 1);