# migrations

## Setup

Перед запуском сервера необходимо выполнить миграцию базы данных.
Предполагается что при развертывании окружения это будет выполнено автоматически.

Формат именования миграций:
./migrations/<id>_<some_name>_<up_or_down>.sql

Запуск миграций:
task migrate
или
go run ./cmd/migrator --migrations-path=./migrations