-- auto-generated definition
create database jobs;

create table jobs
(
  id           serial       not null,
  guid         varchar(128) not null
    constraint jobs_pk
    primary key,
  profile           varchar(128) not null,
  c24_job_id        varchar(128) not null,
  action            varchar(128) not null,
  metadata JSONB,
  created_date timestamp default CURRENT_TIMESTAMP,
  status       varchar(64)
);

alter table jobs
  owner to postgres;

create unique index jobs_id_uindex
  on jobs (id);

create unique index jobs_guid_uindex
  on jobs (guid);

create index jobs_status_index
  on jobs (status);

-- auto-generated definition
create table transcode
(
  id       serial not null
    constraint transcode_pkey
    primary key,
  data     json,
  progress double precision default 0,
  job_id   integer
    constraint transcode_jobs_id_fk
    references jobs (id)
);

alter table transcode
  owner to postgres;

create unique index transcode_id_uindex
  on transcode (id);

