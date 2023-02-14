SHELL := /bin/bash
# gfd

GOCMD=go
GOMOD=$(GOCMD) mod
GOCLEAN=$(GOCMD) clean
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install

# TODO: Указать им сервиса
SERVICE_NAME?=go-service-template
SERVICE_PATH?=$(HOME)
SERVICE_HOME=$(SERVICE_PATH)/$(SERVICE_NAME)

DATABASE_URI=$(URL)
DATABASE_MIGRATIONS_PATH=$(MIGRATIONS_PATH)

LINTER_RESULT_CODE?=1

SWAGGER_VER?=v1.8.6

.PHONY: clean
clean:
	echo "Task: clean."
	$(GOCLEAN)
	rm -f $(SERVICE_NAME)

.PHONY: build
build:
	echo "Task: build. Загрузка зависимостей."
	$(GOMOD) download
	echo "Task: build. Проверка наличия зависимостей."
	$(GOMOD) tidy
	echo "Task: build. Проверка зависимостей текущего модуля."
	$(GOMOD) verify

	make swagger
	$(GOBUILD) -o $(SERVICE_HOME)/$(SERVICE_NAME) ./cmd/main.go

.PHONY: swagger
swagger:
	echo "Task: swagger. Генерация документации Swagger."
	which swag || $(GOINSTALL) github.com/swaggo/swag/cmd/swag@"${SWAGGER_VER}"
	swag init -g ./internal/app/main.go --output docs/swagger

.PHONY: unit-test
unit-test:
	echo "Task. unit-test. Очистка кэша."
	$(GOCLEAN) -cache
	echo "Task. unit-test. Начинаю запуск юнит тестов."
	gotestsum --junitfile test-results/unit-test-report.xml --format testname -- -short ./...

.PHONY: integration-test
integration-test:
	echo "Task. integration-test."
	echo "Task. integration-test. Очистка кэша."
	$(GOCLEAN) -cache
	echo "Task. integration-test. Начинаю запуск интеграционных тестов."
	gotestsum --junitfile test-results/integration-test-report.xml --format testname -- -run TestIntegration ./...

.PHONY: analyze
analyze:
	echo "Task. analyze. Запуск анализа кода."
	golangci-lint run --issues-exit-code $(LINTER_RESULT_CODE) --out-format code-climate | tee gl-code-quality-report.json | jq -r '.[] | "\(.location.path):\(.location.lines.begin) \(.description)"'

.PHONY: database
database:
	echo "Task. database. Выполнение миграций базы данных."
	migrate -path .${DATABASE_MIGRATIONS_PATH} -database "${DATABASE_URI}" -verbose up