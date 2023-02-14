BEGIN;
alter table kafka_out_error_messages alter column timestamp_pc type timestamptz;
alter table kafka_out_error_messages alter column send_time type timestamptz;
COMMIT;