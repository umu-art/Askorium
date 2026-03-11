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

create unique index if not exists uk_sync_tasks_active on sync_tasks (source_id) where status = 'RUNNING';

create table if not exists sync_task_urls
(
    id      uuid                        default uuid_generate_v4() not null,
    created timestamp(6) with time zone default now()              not null,
    updated timestamp(6) with time zone,
    task_id uuid                                                   not null,
    url     varchar(2048)                                          not null,
    status  varchar(20)                                            not null,
    constraint pk_sync_task_urls primary key (id),
    constraint fk_sync_task_urls_on_task_id foreign key (task_id) references sync_tasks on delete cascade,
    constraint uk_sync_task_urls_task_url unique (task_id, url)
);

create index if not exists idx_sync_task_urls_task_id on sync_task_urls (task_id);
