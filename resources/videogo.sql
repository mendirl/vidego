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




select *
from videogo.video
where duration in (select duration from videogo.video group by duration having count(1) > 1)
order by duration desc;