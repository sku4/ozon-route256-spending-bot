-- +goose Up
-- +goose StatementBegin
insert into event(category_id, event_at, price)
values (1, '2022-10-09', 100000),
       (1, '2022-10-09', 150000),
       (2, '2022-10-08', 5680000),
       (2, '2022-10-08', 57601300);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
truncate table event;
-- +goose StatementEnd
