CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

CREATE TABLE IF NOT EXISTS accounts (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    owner text NOT NULL,
    balance bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO accounts (owner, balance)
SELECT 'account-' || value, value * 10
FROM generate_series(1, 10000) AS value;

-- Prime the exact statement shapes used by workload.sql so they are present
-- when sen builds its synthetic query graph.
SELECT balance FROM accounts WHERE id = (random() * 9999)::bigint + 1;
UPDATE accounts
SET balance = balance + 1, updated_at = now()
WHERE id = (random() * 9999)::bigint + 1;
