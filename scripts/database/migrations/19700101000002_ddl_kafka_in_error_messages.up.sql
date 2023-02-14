BEGIN;

CREATE SEQUENCE if not exists kafka_in_error_messages_sq
    START WITH 1 INCREMENT BY 1;

create table if not exists kafka_in_error_messages
(
    id                 int8      not null,

    headers_cs         jsonb,
    timestamp_cs       timestamp,
    block_timestamp_cs timestamp,

    key_cs             bytea,
    value_cs           bytea     not null,
    topic_cs           varchar   not null,
    partition_cs       int4,
    offset_cs          int8,

    error_text         varchar   not null,
    receive_time       timestamp not null
);

comment on table kafka_in_error_messages is 'Incoming message processing errors history';
comment on column kafka_in_error_messages.id is 'ID';

comment on column kafka_in_error_messages.headers_cs is 'Message headers';
comment on column kafka_in_error_messages.timestamp_cs is 'message timestamp';
comment on column kafka_in_error_messages.block_timestamp_cs is 'block timestamp';

comment on column kafka_in_error_messages.key_cs is 'message key';
comment on column kafka_in_error_messages.value_cs is 'message';
comment on column kafka_in_error_messages.topic_cs is 'topic name';
comment on column kafka_in_error_messages.partition_cs is 'partition';
comment on column kafka_in_error_messages.offset_cs is 'message offset';

comment on column kafka_in_error_messages.error_text is 'error text';
comment on column kafka_in_error_messages.receive_time is 'event timestamp';

COMMIT;