BEGIN;

DROP FUNCTION IF EXISTS schema_migrations_history_add_schema_migrations_fnc() CASCADE;
DROP TRIGGER IF EXISTS schema_migrations_insert_or_update_trigger ON schema_migrations CASCADE;

COMMIT;