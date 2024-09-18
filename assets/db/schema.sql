create table tender
(
    uid         integer           not null
        constraint tender_pk
            primary key autoincrement,
    uuid        text              not null
        constraint tender_pk2
            unique,
    created_at  integer default 0 not null,
    updated_at  integer default 0 not null,
    source      text              not null,
    link        text              not null
        constraint tender_pk3
            unique,
    title       text              not null,
    unix_date   integer default 0 not null,
    categories  text              not null,
    description text              not null,
    guid        text              not null,
    json_version text              not null,
    json        text              not null,
    attempt integer default 0 not null,
    error text not null,
    publish_at integer default 0 not null,
    closing_at integer default 0 not null,
    deleted integer default 0 not null
);