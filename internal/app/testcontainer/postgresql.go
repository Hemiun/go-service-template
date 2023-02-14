package testcontainer

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/go-playground/validator/v10"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4"
	"github.com/testcontainers/testcontainers-go"
	"go-service-template/internal/app/infrastructure"
)

//--------------

const (
	createDatabasePath = "scripts/database/init/01.database.sql"
	createSchemaPath   = "scripts/database/init/02.schema.sql"
	migrationPath      = "scripts/database/migrations"
)

func buildPath(pathFromRoot string) string {
	currentPath, _ := os.Getwd()
	var res string
	for _, elem := range strings.Split(currentPath, string(filepath.Separator)) {
		if elem != "internal" {
			res = path.Join(res, elem)
		} else {
			break
		}
	}
	res = path.Join(res, pathFromRoot)
	res = filepath.ToSlash(res)
	if runtime.GOOS != "windows" {
		res = "/" + res
	}

	return res
}

// DatabaseContainerConfig configuration for database container (postgresql)
type DatabaseContainerConfig struct {
	DatabaseName      string        `validate:"required"`
	OwnerSchema       string        `validate:"required"`
	OwnerSchemaPass   string        `validate:"required"`
	ServiceSchema     string        `validate:"required"`
	ServiceSchemaPass string        `validate:"required"`
	Timeout           time.Duration `validate:"required"`
	Waiting           time.Duration
}

// Validate - validation for DatabaseContainerConfig
func (c *DatabaseContainerConfig) Validate() error {
	if c.Waiting == 0 {
		c.Waiting = time.Minute * 1
	}
	return validator.New().Struct(c)
}

// DatabaseContainer - struct for db container
type DatabaseContainer struct {
	infrastructure.SugarLogger
	instance testcontainers.Container
	cfg      DatabaseContainerConfig
}

// NewDatabaseContainer  returns new DatabaseContainer
func NewDatabaseContainer(ctx context.Context, cfg DatabaseContainerConfig) (*DatabaseContainer, error) {
	var target DatabaseContainer

	testcontainers.Logger = &Logger{}
	target.cfg = cfg
	if err := target.cfg.Validate(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	w := wait.ForSQL("5432/tcp", "postgres", func(port nat.Port) string {
		return fmt.Sprintf("postgres://postgres:postgres@127.0.0.1:%s/postgres?sslmode=disable", port.Port())
	}).WithQuery("select 10")

	req := testcontainers.ContainerRequest{
		Image:        "postgres:14.2",
		ExposedPorts: []string{"5432/tcp"},
		AutoRemove:   true,
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "postgres",
		},
		WaitingFor: w,
	}
	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic("can't start container")
	}

	target.instance = postgres
	return &target, nil
}

// Port - return port for db connection (looking for port mapped into default postgresql port - 5432 )
func (db *DatabaseContainer) Port(ctx context.Context) int {
	ctx, cancel := context.WithTimeout(ctx, db.cfg.Timeout)
	defer cancel()
	p, err := db.instance.MappedPort(ctx, "5432")
	if err != nil {
		db.LogError(ctx, "can't get port", err)
		return 0
	}
	return p.Int()
}

// ConnectionString - returns connection string in dsn format
func (db *DatabaseContainer) ConnectionString(ctx context.Context) string {
	return db.connectionString(ctx, db.cfg.DatabaseName, db.cfg.ServiceSchema, db.cfg.ServiceSchemaPass)
}

func (db *DatabaseContainer) connectionString(ctx context.Context, dbName string, user string, pass string) string {
	return fmt.Sprintf("postgres://%s:%s@127.0.0.1:%d/%s?sslmode=disable", user, pass, db.Port(ctx), dbName)
}

// Close - close created container
func (db *DatabaseContainer) Close(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, db.cfg.Timeout)
	defer cancel()
	err := db.instance.Terminate(ctx)
	if err != nil {
		db.LogError(ctx, "can't close container", err)
	}
}

// PrepareDB  - prepare database structure (new db, new schema, applying migration)
func (db *DatabaseContainer) PrepareDB(ctx context.Context) error {
	err := db.createDBAndSchema(ctx)
	if err != nil {
		db.LogError(ctx, "error while createDBAndSchema", err)
		return err
	}
	err = db.runMigrate(ctx)
	if err != nil {
		db.LogError(ctx, "error while migration apply", err)
		return err
	}
	return nil
}

func (db *DatabaseContainer) buildScriptFromTemplate(ctx context.Context, path string) (string, error) {
	scriptPath := buildPath(path)
	db.LogDebug(ctx, "looking fot a script:"+scriptPath)
	b, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		db.LogError(ctx, "Can't read script py path:"+path, err)
		return "", err
	}
	source := string(b)
	scriptTemplate, err := template.New("db").Parse(source)
	if err != nil {
		db.LogError(ctx, "Can't parse template", err)
		return "", err
	}
	buf := new(bytes.Buffer)
	err = scriptTemplate.Execute(buf, db.cfg)
	if err != nil {
		db.LogError(ctx, "Can't execute template", err)
		return "", err
	}

	return buf.String(), nil
}

func (db *DatabaseContainer) createDBAndSchema(ctx context.Context) error {
	db.LogDebug(ctx, "container_id:"+db.instance.GetContainerID())
	ctx, cancel := context.WithTimeout(ctx, db.cfg.Timeout)
	defer cancel()
	dsn := db.connectionString(ctx, "postgres", "postgres", "postgres")
	db.LogDebug(ctx, "dsn:"+dsn)
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		db.LogPanic(ctx, "Can't get connection", err)
	}

	// db creation
	script, err := db.buildScriptFromTemplate(ctx, createDatabasePath)
	if err != nil {
		db.LogPanic(ctx, "Can't build script from template (create database)", err)
	}
	_, err = conn.Exec(ctx, script)
	if err != nil {
		db.LogPanic(ctx, "Can't execute script:"+createDatabasePath, err)
	}
	//  reconnect to created db
	err = conn.Close(ctx)
	if err != nil {
		db.LogError(ctx, "can't close connection", err)
	}

	// creates user and schema
	dsn = db.connectionString(ctx, db.cfg.DatabaseName, "postgres", "postgres")
	conn, err = pgx.Connect(ctx, dsn)
	defer func(conn *pgx.Conn, ctx context.Context) {
		e := conn.Close(ctx)
		if e != nil {
			db.LogError(ctx, "can't close connection", e)
		}
	}(conn, ctx)
	if err != nil {
		db.LogPanic(ctx, "Can't get connection", err)
	}

	script, err = db.buildScriptFromTemplate(ctx, createSchemaPath)
	if err != nil {
		db.LogPanic(ctx, "Can't build script from template (create schema)", err)
	}
	db.LogDebug(ctx, script)
	_, err = conn.Exec(ctx, script)
	if err != nil {
		db.LogPanic(ctx, "Can't execute script:"+createSchemaPath, err)
	}

	return nil
}

func (db *DatabaseContainer) runMigrate(ctx context.Context) error {
	dsn := db.connectionString(ctx, db.cfg.DatabaseName, db.cfg.OwnerSchema, db.cfg.OwnerSchemaPass)

	m, err := migrate.New("file://"+path.Clean(buildPath(migrationPath)), dsn)
	if err != nil {
		db.LogError(ctx, "can't create migrate struct", err)
		return err
	}
	err = m.Up()
	if err != nil {
		db.LogError(ctx, "can't apply migrates", err)
		return err
	}
	return nil
}
