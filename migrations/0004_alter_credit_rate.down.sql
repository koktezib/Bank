-- migrations/0004_alter_credit_rate.down.sql
ALTER TABLE credits
ALTER COLUMN annual_rate TYPE NUMERIC(5,4)
    USING ROUND(annual_rate::numeric, 4)::numeric(5,4);
