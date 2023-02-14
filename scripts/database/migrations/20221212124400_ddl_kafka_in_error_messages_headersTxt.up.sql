BEGIN;
    alter table kafka_in_error_messages add column headers_txt  bytea;
    comment on column kafka_in_error_messages.headers_txt is 'Message headers in text form';
COMMIT;