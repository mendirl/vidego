drop schema if exists vidego cascade;
create schema if not exists vidego;

drop table if exists vidego.video;
create table vidego.video
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

drop table if exists vidego.config;
create table vidego.config
(
    position serial,
    name     text unique,
    values   text[]
);

alter table vidego.video
    add column to_delete boolean default false;

---------------

truncate table vidego.video;

-- count

select count(1)
from vidego.video;
select path,
       count(1)
from vidego.video
group by path
order by path;


-- dedup

-- liste des elements de video qui ont la meme size et duration

select *
from vidego.video
where path not like '%dedup%'
                        and deleted_at is null
                        and duration in
                            (select duration from vidego.video where path not like '%dedup' and deleted_at is null group by duration having count(1) > 1)
order by duration desc;

select *
from vidego.video
where (size, duration) in (select size, duration
                           from vidego.video
                           where name is not null
                             and name <> ''
                           group by size, duration
                           having count(*) > 1)
  and name is not null
  and name <> ''
  and to_delete is false
order by duration desc;
-- putback

select *
from vidego.video
where deduplicate is true
order by duration desc


select * from vidego.video where path like '%dedup%' and deleted_at is null;

update vidego.video set deduplicate = true where path like '%dedup%' and deleted_at is null;


-- delete


select *
from vidego.video
where to_delete is true;




