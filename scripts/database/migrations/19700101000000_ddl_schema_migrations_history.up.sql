BEGIN;
CREATE SEQUENCE if not exists schema_migrations_history_sq
    START WITH 1 INCREMENT BY 1;

create table if not exists schema_migrations_history
(
    id              bigint                  PRIMARY KEY,
    update_time     timestamp               NOT NULL,
    operation_type  varchar(50)             NOT NULL,
    version         bigint                  NOT NULL,
    dirty           boolean                 NOT NULL
);

comment on table schema_migrations_history is 'История миграций';
comment on column schema_migrations_history.id is 'Идентификатор записи';
comment on column schema_migrations_history.operation_type is 'Тип операции';
comment on column schema_migrations_history.update_time is 'Время миграции';
comment on column schema_migrations_history.version is 'Версия миграции';
comment on column schema_migrations_history.dirty is 'Миграция не выполнена';

COMMIT;