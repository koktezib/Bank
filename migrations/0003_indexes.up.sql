-- migrations/0003_indexes.up.sql

CREATE INDEX ON transactions(account_id);
CREATE INDEX ON payment_schedules(credit_id);
