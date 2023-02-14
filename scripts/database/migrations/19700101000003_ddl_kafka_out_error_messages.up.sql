BEGIN;


CREATE SEQUENCE if not exists kafka_out_error_messages_sq
    START WITH 1 INCREMENT BY 1;

create table if not exists kafka_out_error_messages
(
    id                int8      not null,

    topic_pc          varchar   not null,
    key_pc            bytea,
    value_pc          bytea     not null,
    headers_pc        jsonb,

    metadata_pc       jsonb,
    offset_pc         int8,
    partition_pc      int4,
    timestamp_pc      timestamp,

    error_text        varchar   not null,
    send_time         timestamp not null
);

comment on table kafka_out_error_messages is 'Outgoing message processing errors history';
comment on column kafka_out_error_messages.id is 'ID';

comment on column kafka_out_error_messages.topic_pc is 'Topic';
comment on column kafka_out_error_messages.key_pc is 'Message key';
comment on column kafka_out_error_messages.value_pc is 'Message';
comment on column kafka_out_error_messages.headers_pc is 'Message headers';

comment on column kafka_out_error_messages.metadata_pc is 'Metadata';
comment on column kafka_out_error_messages.offset_pc is 'Message offset';
comment on column kafka_out_error_messages.partition_pc is 'partition';
comment on column kafka_out_error_messages.timestamp_pc is 'message timestamp';

comment on column kafka_out_error_messages.error_text is 'error text';
comment on column kafka_out_error_messages.send_time is 'event timestamp';

COMMIT;
