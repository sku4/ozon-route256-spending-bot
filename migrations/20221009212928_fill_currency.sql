-- +goose Up
-- +goose StatementBegin
insert into currency(abbreviation)
values ('USD'),('CNY'),('EUR'),('RUB');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
truncate currency cascade;
-- +goose StatementEnd
