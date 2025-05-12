# Bank API

## Оглавление

* [Требования](#требования)
* [Установка](#установка)
* [Конфигурация](#конфигурация)
* [Миграции базы данных](#миграции-базы-данных)
* [Запуск сервиса](#запуск-сервиса)
* [Команды](#команды)
* [API Endpoints](#api-endpoints)
* [Примеры запросов](#примеры-запросов)
* [Тестирование](#тестирование)

## Требования

* Go 1.23+
* PostgreSQL 17 с расширением `pgcrypto`
* Git

## Установка

1. Клонируйте репозиторий:

   ```bash
   git clone https://github.com/ваш/login/bank-api.git
   cd bank-api
   ```
2. Установите зависимости Go:

   ```bash
   go mod download
   ```

## Конфигурация

1. Скопируйте пример переменных окружения:

   ```bash
   cp .env.example .env
   ```
2. Откройте `.env` и заполните:

   ```ini
   # Database
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASS=пароль
   DB_NAME=bank

   # JWT
   JWT_SECRET=ваш_секрет

   # SMTP (Gmail)
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USER=your_account@gmail.com
   SMTP_PASS=app_password

   # PGP ключи
   PGP_PUBLIC_KEY=/path/to/pubkey.asc
   PGP_PRIVATE_KEY=/path/to/privkey.asc
   PGP_PASSPHRASE=coca-cola

   # HMAC
   HMAC_SECRET=ваш_hmac_секрет
   ```

## Миграции базы данных

Все SQL-скрипты находятся в папке `migrations/`.

Установите CLI миграций:

```bash
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Примените миграции:

```bash
migrate -path ./migrations -database "postgres://$DB_USER:$DB_PASS@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" up
```

Для отката одной миграции:

```bash
migrate -path ./migrations -database "..." down 1
```

## Запуск сервиса

```bash
go run cmd/server/main.go
```

По умолчанию слушает `:8080`.

## Команды

* `go run cmd/server/main.go` — запуск приложения
* `migrate up` / `migrate down` — управление миграциями
* `go test ./...` — запуск тестов

## API Endpoints

### Public

* `POST /register` — регистрация пользователя
* `POST /login` — получение JWT

### Protected (Bearer JWT)

* `POST   /accounts` — создать счёт
* `GET    /accounts` — список счётов
* `POST   /accounts/deposit` — пополнение счёта
* `POST   /accounts/withdraw` — снятие средств
* `POST   /transfer` — перевод между счетами
* `POST   /cards` — выпустить карту (query: `?account_id=`)
* `GET    /cards` — список карт
* `POST   /credits` — оформление кредита
* `GET    /credits/{creditId}/schedule` — график платежей по кредиту
* `GET    /analytics` — статистика доходов/расходов/кредитной нагрузки
* `GET    /accounts/{accountId}/predict?days=N` — прогноз баланса на N дней

## Примеры запросов

### Регистрация

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","email":"user1@example.com","password":"pass123"}'
```

### Аутентификация

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user1@example.com","password":"pass123"}'
```

### Создание счёта

```bash
curl -X POST http://localhost:8080/accounts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"currency":"RUB"}'
```
