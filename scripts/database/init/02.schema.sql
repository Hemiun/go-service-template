-- 1. Commands for Database users creation. First user - schema owner, second - user for service
CREATE USER {{ .OwnerSchema}} WITH password '{{ .OwnerSchemaPass}}';
CREATE USER {{ .ServiceSchema}} WITH password '{{ .ServiceSchemaPass}}';

-- 2. Schema creation
CREATE SCHEMA {{ .OwnerSchema}}  AUTHORIZATION {{ .OwnerSchema}};

-- 3. Grant
GRANT USAGE ON SCHEMA {{ .OwnerSchema}} TO {{ .ServiceSchema}};

ALTER DEFAULT PRIVILEGES FOR USER {{ .OwnerSchema}} IN SCHEMA {{ .OwnerSchema}}  GRANT SELECT,INSERT,UPDATE,DELETE,TRUNCATE ON TABLES TO {{ .ServiceSchema}};
ALTER DEFAULT PRIVILEGES FOR USER {{ .OwnerSchema}} IN SCHEMA {{ .OwnerSchema}}  GRANT USAGE ON SEQUENCES TO {{ .ServiceSchema}};
ALTER DEFAULT PRIVILEGES FOR USER {{ .OwnerSchema}} IN SCHEMA {{ .OwnerSchema}}  GRANT EXECUTE ON FUNCTIONS TO {{ .ServiceSchema}};

alter role {{ .ServiceSchema}} set search_path = '{{ .OwnerSchema}}';