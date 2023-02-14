BEGIN;
    alter table kafka_out_error_messages drop column headers_txt;
COMMIT;