BEGIN;
alter table kafka_in_error_messages alter column timestamp_cs type timestamptz;
alter table kafka_in_error_messages alter column block_timestamp_cs type timestamptz;
alter table kafka_in_error_messages alter column receive_time type timestamptz;
COMMIT;