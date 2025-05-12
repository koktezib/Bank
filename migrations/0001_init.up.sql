-- migrations/0001_init.up.sql

-- подключаем pgcrypto для шифрования PGP
CREATE EXTENSION IF NOT EXISTS pgcrypto;
