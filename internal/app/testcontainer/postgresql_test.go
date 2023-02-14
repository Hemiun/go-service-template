package testcontainer

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-playground/validator/v10"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseContainer_buildScriptFromTemplate(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		cfg     DatabaseContainerConfig
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Case 1. Positive(create database)",
			cfg: DatabaseContainerConfig{
				DatabaseName:      "territory",
				OwnerSchema:       "test",
				OwnerSchemaPass:   "test",
				ServiceSchema:     "test_ms",
				ServiceSchemaPass: "test_ms",
				Timeout:           time.Minute,
			},
			args: args{
				path: createDatabasePath,
			},
			wantErr: false,
		},
		{
			name: "Case 2. Bad script name",
			cfg: DatabaseContainerConfig{
				DatabaseName:      "territory",
				OwnerSchema:       "test",
				OwnerSchemaPass:   "test",
				ServiceSchema:     "test_ms",
				ServiceSchemaPass: "test_ms",
				Timeout:           time.Minute,
			},
			args: args{
				path: "dsfds",
			},
			wantErr: true,
		},
		{
			name: "Case 3. Empty script name",
			cfg: DatabaseContainerConfig{
				DatabaseName:      "territory",
				OwnerSchema:       "test",
				OwnerSchemaPass:   "test",
				ServiceSchema:     "test_ms",
				ServiceSchemaPass: "test_ms",
				Timeout:           time.Minute,
			},
			args: args{
				path: "",
			},
			wantErr: true,
		},
		{
			name: "Case 4. Positive(create schemas)",
			cfg: DatabaseContainerConfig{
				DatabaseName:      "territory",
				OwnerSchema:       "test",
				OwnerSchemaPass:   "test",
				ServiceSchema:     "test_ms",
				ServiceSchemaPass: "test_ms",
				Timeout:           time.Minute,
			},
			args: args{
				path: createSchemaPath,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := &DatabaseContainer{
				cfg: tt.cfg,
			}
			got, err := target.buildScriptFromTemplate(context.Background(), tt.args.path)
			fmt.Println(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildScriptFromTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestIntegrationDatabaseContainer_createDBAndSchema(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		cfg     DatabaseContainerConfig
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Case 1. Positive(create database)",
			cfg: DatabaseContainerConfig{
				DatabaseName:      "territory",
				OwnerSchema:       "test",
				OwnerSchemaPass:   "test",
				ServiceSchema:     "test_ms",
				ServiceSchemaPass: "test_ms",
				Timeout:           time.Minute,
			},
			args: args{
				path: createDatabasePath,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := NewDatabaseContainer(context.Background(), tt.cfg)
			assert.NoError(t, err)
			err = target.createDBAndSchema(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("buildScriptFromTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestIntegrationDatabaseContainer_PrepareDB(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		cfg     DatabaseContainerConfig
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Case 1. Positive(prepare database)",
			cfg: DatabaseContainerConfig{
				DatabaseName:      "territory",
				OwnerSchema:       "test",
				OwnerSchemaPass:   "test",
				ServiceSchema:     "test_ms",
				ServiceSchemaPass: "test_ms",
				Timeout:           time.Minute,
			},
			args: args{
				path: createDatabasePath,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := context.Background()
			target, err := NewDatabaseContainer(c, tt.cfg)
			assert.NoError(t, err)
			err = target.PrepareDB(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildScriptFromTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_buildPath(t *testing.T) {
	tests := []struct {
		name         string
		pathFromRoot string
		want         string
	}{
		{
			name: "Case 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := buildPath(migrationPath)
			fmt.Println(path)
			file, err := os.Open(path)
			assert.NoError(t, err)
			defer file.Close() //nolint:errcheck
		})
	}
}
