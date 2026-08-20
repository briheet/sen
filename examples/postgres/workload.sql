BEGIN;
SELECT balance FROM accounts WHERE id = (random() * 9999)::bigint + 1;
UPDATE accounts
SET balance = balance + 1, updated_at = now()
WHERE id = (random() * 9999)::bigint + 1;
COMMIT;
