-- migrations/0004_alter_credit_rate.up.sql
ALTER TABLE credits
ALTER COLUMN annual_rate TYPE NUMERIC(6,4)
    USING ROUND(annual_rate::numeric, 4)::numeric(6,4);
