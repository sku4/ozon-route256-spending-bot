-- +goose Up
-- +goose StatementBegin
insert into category(title)
values ('Food'),
       ('Auto'),
       ('Medicine'),
       ('Entertainment');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
truncate category cascade;
-- +goose StatementEnd
