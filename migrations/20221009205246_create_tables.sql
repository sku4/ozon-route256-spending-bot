-- +goose Up
-- +goose StatementBegin
create table currency
(
    id           int generated always as identity,
    abbreviation varchar(3) not null,
    primary key (id)
);

create table rate
(
    id          int generated always as identity,
    currency_id int,
    rate        bigint,
    primary key (id),
    constraint fk_rate_currency
        foreign key (currency_id)
            references currency (id)
            on delete cascade
);

create table state
(
    id          int generated always as identity,
    currency_id int,
    primary key (id),
    constraint fk_state_currency
        foreign key (currency_id)
            references currency (id)
            on delete set null
);

create table "user"
(
    id          int generated always as identity,
    telegram_id int,
    state_id    int,
    primary key (id),
    constraint fk_user_state
        foreign key (state_id)
            references state (id)
            on delete set null
);

create table category
(
    id         int generated always as identity,
    title      varchar(255) not null,
    created_at timestamp    not null default now(),
    primary key (id)
);

create table event
(
    id          int generated always as identity,
    category_id int,
    event_at    date      not null,
    price       bigint,
    created_at  timestamp not null default now(),
    primary key (id),
    constraint fk_event_category
        foreign key (category_id)
            references category (id)
            on delete cascade
);

create table category_limit
(
    id             int generated always as identity,
    state_id       int references state (id) on delete cascade,
    category_id    int references category (id) on delete cascade,
    category_limit bigint,
    primary key (id)
);

-- index by date because report build only by range dates
create index event_at_idx on event (event_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table "user";
drop table state;
drop table rate;
drop table currency;
drop table event;
drop table category;
drop table category_limit;
-- +goose StatementEnd
