# Шаблон сервиса на go
Задает структуру проекта.  
Реализует метод для опроса состояния(БД, kafka).  
Работа со Swagger.  
Конфигурирование.   
Логирование.

## [Как работать с шаблоном (HOW TO)](./docs/howto.md)
## [Технологический стек](./docs/tech.md)

## swagger

### [Swagger-ui](http://localhost:8080/swagger-ui/index.html)

При необходимости заменить адрес и порт. Протокол - http

### Генерация документации

`swag init -g .\internal\app\main.go --output docs/swagger`  
Сгенерированные файлы должны быть добавлены в репозиторий

Upd: Вызов swag из корня почему-то перестал работать. WA:
> cd ./internal/app
> swag init  --output ../../docs/swagger