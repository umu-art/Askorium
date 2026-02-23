create extension if not exists "uuid-ossp";

create table if not exists users
(
    id                     uuid                        default uuid_generate_v4() not null,
    created                timestamp(6) with time zone default now()              not null,
    updated                timestamp(6) with time zone,
    last_seen_at           timestamp(6) with time zone,
    last_seen_ip           varchar(64),
    first_visit_user_agent varchar(1024),
    first_visit_headers    text,
    constraint pk_users primary key (id)
);

create table if not exists sources
(
    id         uuid                        default uuid_generate_v4() not null,
    created    timestamp(6) with time zone default now()              not null,
    updated    timestamp(6) with time zone,
    source_url varchar(2048)                                          not null,
    constraint pk_sources primary key (id),
    constraint uk_sources_user_id_source_url unique (source_url)
);

create table if not exists source_sync_policies
(
    id               uuid                        default uuid_generate_v4() not null,
    created          timestamp(6) with time zone default now()              not null,
    updated          timestamp(6) with time zone,
    source_id        uuid                                                   not null,
    enabled          boolean                     default true               not null,
    interval_minutes integer                     default 720                not null,
    last_synced_at   timestamp(6) with time zone,
    constraint pk_source_sync_policies primary key (id),
    constraint fk_source_sync_policies_on_source_id foreign key (source_id) references sources on delete cascade
);

create table if not exists pages
(
    id           uuid                        default uuid_generate_v4() not null,
    created      timestamp(6) with time zone default now()              not null,
    updated      timestamp(6) with time zone,
    source_id    uuid                                                   not null,
    url          varchar(2048)                                          not null,
    title        varchar(1024)                                          not null,
    preview_url  varchar(2048),
    icon_url     varchar(2048),
    description  text,
    language     varchar(10),
    content_hash varchar(128),
    constraint pk_pages primary key (id),
    constraint fk_pages_on_source_id foreign key (source_id) references sources on delete cascade,
    constraint uk_pages_url unique (url)
);

create table if not exists page_blocks
(
    id            uuid                        default uuid_generate_v4() not null,
    created       timestamp(6) with time zone default now()              not null,
    updated       timestamp(6) with time zone,
    page_id       uuid                                                   not null,
    html_id       varchar(255),
    type          varchar(50)                                            not null,
    heading_level integer,
    text          text                                                   not null,
    constraint pk_blocks primary key (id),
    constraint fk_blocks_on_page_id foreign key (page_id) references pages on delete cascade,
    constraint uk_blocks_page_id_html_id unique (page_id, html_id),
    constraint uk_blocks_page_id_text unique (page_id, text)
);

create table if not exists page_links
(
    id         uuid                        default uuid_generate_v4() not null,
    created    timestamp(6) with time zone default now()              not null,
    updated    timestamp(6) with time zone,
    page_id    uuid                                                   not null,
    block_id   uuid,
    href       varchar(2048)                                          not null,
    type       varchar(20)                                            not null,
    anchorText text,
    snippet    text,
    position   integer                                                not null,
    constraint pk_page_links primary key (id),
    constraint fk_page_links_on_page_id foreign key (page_id) references pages on delete cascade,
    constraint fk_page_links_on_block_id foreign key (block_id) references page_blocks on delete set null
);

create table if not exists page_documents
(
    id                 uuid                        default uuid_generate_v4() not null,
    created            timestamp(6) with time zone default now()              not null,
    updated            timestamp(6) with time zone,
    page_id            uuid                                                   not null,
    url                varchar(2048)                                          not null,
    mime_type          varchar(255)                                           not null,
    size_bytes         integer,
    extracted_text     text,
    description        text,
    description_source varchar(255),
    constraint pk_page_documents primary key (id),
    constraint fk_page_documents_on_page_id foreign key (page_id) references pages on delete cascade
);

create table if not exists search_queries
(
    id          uuid                        default uuid_generate_v4() not null,
    created     timestamp(6) with time zone default now()              not null,
    updated     timestamp(6) with time zone,
    user_id     uuid                                                   not null,
    source_id   uuid                                                   not null,
    status      varchar(20)                                            not null,
    query            text                                                   not null,
    normalized_query text,
    query_vector     jsonb,
    mode             varchar(20)                                            not null,
    answer      text,
    error       text,
    finished_at timestamp(6) with time zone,
    constraint pk_search_queries primary key (id),
    constraint fk_search_queries_on_user_id foreign key (user_id) references users on delete cascade
);

create table if not exists search_query_sources
(
    id           uuid                        default uuid_generate_v4() not null,
    created      timestamp(6) with time zone default now()              not null,
    updated      timestamp(6) with time zone,
    query_id     uuid                                                   not null,
    index_key    varchar(128)                                           not null,
    url          varchar(2048)                                          not null,
    title        varchar(1024)                                          not null,
    date         timestamp(6) with time zone,
    text         text                                                   not null,
    score_sparse real,
    score_dense  real,
    fusion_score real,
    score_final  real,
    constraint pk_search_query_sources primary key (id),
    constraint fk_search_query_sources_on_query_id foreign key (query_id) references search_queries on delete cascade,
    constraint uk_search_query_sources_query_id_block_id unique (query_id, index_key)
);

create table if not exists feedbacks
(
    id       uuid                        default uuid_generate_v4() not null,
    created  timestamp(6) with time zone default now()              not null,
    updated  timestamp(6) with time zone,
    query_id uuid                                                   not null,
    user_id  uuid                                                   not null,
    rating   integer                                                not null,
    text     text,
    constraint pk_feedbacks primary key (id),
    constraint fk_feedbacks_on_query_id foreign key (query_id) references search_queries on delete cascade,
    constraint fk_feedbacks_on_user_id foreign key (user_id) references users on delete cascade,
    constraint feedbacks_rating_check check (rating >= 1 and rating <= 5)
);

create index if not exists idx_pages_source_id on pages (source_id);

create index if not exists idx_blocks_page_id on page_blocks (page_id);

create index if not exists idx_page_links_page_id on page_links (page_id);

create index if not exists idx_page_documents_page_id on page_documents (page_id);

create index if not exists idx_search_queries_user_id on search_queries (user_id);
create index if not exists idx_search_queries_source_id on search_queries (source_id);
create index if not exists idx_search_queries_status on search_queries (status);

create index if not exists idx_search_query_sources_query_id on search_query_sources (query_id);

create index if not exists idx_feedbacks_query_id on feedbacks (query_id);
