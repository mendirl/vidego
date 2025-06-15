drop schema if exists videogo cascade;
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

drop table if exists videogo.config;
create table videogo.config
(
    position serial,
    name     text unique,
    values   text[]
);


---------------

truncate table videogo.video;


-- putback

select *
from videogo.video
where deduplicate is true order by duration desc;
