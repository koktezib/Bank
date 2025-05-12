-- migrations/0003_indexes.down.sql

DROP INDEX IF EXISTS transactions_account_id_idx;
DROP INDEX IF EXISTS payment_schedules_credit_id_idx;
