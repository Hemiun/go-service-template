BEGIN;

CREATE OR REPLACE FUNCTION schema_migrations_history_add_schema_migrations_fnc()
    RETURNS trigger
    LANGUAGE 'plpgsql' AS
$$
DECLARE
    _new_schema_migrations_history_id bigint;
BEGIN
    SELECT INTO _new_schema_migrations_history_id nextval('schema_migrations_history_sq');
    INSERT INTO schema_migrations_history (id, update_time, operation_type, version, dirty)
    VALUES (_new_schema_migrations_history_id, now(), TG_OP, NEW.version, NEW.dirty);
    RETURN NEW;
END
$$;

CREATE OR REPLACE TRIGGER schema_migrations_insert_or_update_trigger
    AFTER INSERT OR UPDATE
    ON schema_migrations
    FOR EACH ROW
EXECUTE PROCEDURE schema_migrations_history_add_schema_migrations_fnc();

COMMIT;