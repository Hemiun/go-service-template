BEGIN;
    alter table kafka_in_error_messages drop column headers_txt;
COMMIT;