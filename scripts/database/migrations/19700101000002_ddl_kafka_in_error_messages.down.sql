BEGIN;
DROP SEQUENCE IF EXISTS kafka_in_error_messages_sq CASCADE;
DROP TABLE IF EXISTS kafka_in_error_messages CASCADE;
COMMIT;