BEGIN;

DROP SEQUENCE IF EXISTS schema_migrations_history_sq CASCADE;
DROP TABLE IF EXISTS schema_migrations_history CASCADE;

COMMIT;