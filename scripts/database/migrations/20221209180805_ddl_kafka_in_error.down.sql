BEGIN;
alter table kafka_in_error_messages alter column timestamp_cs type timestamp;
alter table kafka_in_error_messages alter column block_timestamp_cs type timestamp;
alter table kafka_in_error_messages alter column receive_time type timestamp;
COMMIT;