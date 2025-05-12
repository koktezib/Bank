-- migrations/0002_schema.up.sql

-- 1. Таблица пользователей
CREATE TABLE users (
                       id              SERIAL PRIMARY KEY,
                       username        VARCHAR(50) NOT NULL UNIQUE,
                       email           VARCHAR(100) NOT NULL UNIQUE,
                       password_hash   TEXT    NOT NULL,
                       created_at      TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- 2. Таблица счетов
CREATE TABLE accounts (
                          id          SERIAL PRIMARY KEY,
                          user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                          balance     NUMERIC(18,2) NOT NULL DEFAULT 0,
                          currency    CHAR(3) NOT NULL DEFAULT 'RUB',
                          created_at  TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- 3. Таблица карт
CREATE TABLE cards (
                       id                   SERIAL PRIMARY KEY,
                       account_id           INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
                       number_encrypted     BYTEA   NOT NULL,        -- PGP-шифротекст
                       expiry_encrypted     BYTEA   NOT NULL,        -- PGP-шифротекст
                       cvv_hash             TEXT    NOT NULL,        -- bcrypt
                       hmac                 TEXT    NOT NULL,        -- HMAC-SHA256 от clear-number
                       created_at           TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- 4. История операций
CREATE TABLE transactions (
                              id            SERIAL PRIMARY KEY,
                              account_id    INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
                              amount        NUMERIC(18,2) NOT NULL,
                              type          VARCHAR(20) NOT NULL,  -- e.g. 'deposit','withdraw','transfer'
                              description   TEXT,
                              created_at    TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- 5. Таблица кредитов
CREATE TABLE credits (
                         id             SERIAL PRIMARY KEY,
                         account_id     INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
                         principal      NUMERIC(18,2) NOT NULL,
                         annual_rate    NUMERIC(5,4) NOT NULL,    -- например 0.1234 = 12.34%
                         term_months    INTEGER NOT NULL,
                         created_at     TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- 6. График платежей по кредиту
CREATE TABLE payment_schedules (
                                   id             SERIAL PRIMARY KEY,
                                   credit_id      INTEGER NOT NULL REFERENCES credits(id) ON DELETE CASCADE,
                                   due_date       DATE    NOT NULL,
                                   amount         NUMERIC(18,2) NOT NULL,
                                   paid           BOOLEAN NOT NULL DEFAULT FALSE,
                                   penalty        NUMERIC(18,2) NOT NULL DEFAULT 0,  -- штрафы за просрочку
                                   created_at     TIMESTAMP WITH TIME ZONE DEFAULT now()
);
