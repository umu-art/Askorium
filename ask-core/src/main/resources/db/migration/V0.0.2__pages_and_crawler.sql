create table if not exists sync_tasks
(
    id               uuid                        default uuid_generate_v4() not null,
    created          timestamp(6) with time zone default now()              not null,
    updated          timestamp(6) with time zone,
    source_id        uuid                                                   not null,
    status           varchar(20)                                            not null,
    force_sync       boolean                     default false              not null,
    pages_discovered integer                     default 0                  not null,
    pages_scraped    integer                     default 0                  not null,
    pages_failed     integer                     default 0                  not null,
    error_message    text,
    started_at       timestamp(6) with time zone,
    finished_at      timestamp(6) with time zone,
    constraint pk_sync_tasks primary key (id)
);
