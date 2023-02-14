# syntax=docker/dockerfile:1

FROM golang:1.18.6-bullseye

# Аргументы, приходят из common-ci, номер тега и название гитлаб проекта соответственно.
ARG APP_VER
ARG APP_NAME

ENV GOPATH $GOPATH
ENV SERVICE_VERSION $APP_VER
ENV SERVICE_NAME $APP_NAME
# ENV LOG_LEVEL info

ENV SERVICE_PATH /opt
ENV SERVICE_HOME $SERVICE_PATH/$SERVICE_NAME

LABEL Description="Territory tg adapter" Vendor="I-teco" Version=$SERVICE_VERSION

#  Подготавливаем окружение и права пользователя
RUN apt update \
 && apt upgrade -y \
 && apt install -y --no-install-recommends \
    apt-utils wget curl ca-certificates apt-transport-https gnupg unzip \
 && apt autoremove -y \
 && rm -rf /root/.cache \
 && rm -rf /var/lib/apt/lists/* \
 && rm -rf /var/cache/apt \
 && mkdir -p $SERVICE_HOME \
 && chmod -R 777 $SERVICE_HOME

# Устанавливаем рабочую директорию для сборки проекта
WORKDIR $GOPATH/src/$SERVICE_NAME

# Копируем файлы проекта в докер-контейнер и выполняем сборку проекта
COPY go.mod ./
COPY go.sum ./
COPY ./Makefile ./
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./docs ./docs
COPY ./scripts ./scripts
COPY ./entrypoint.sh $SERVICE_HOME
RUN make build

# Указываем порт, открываемый при запуске приложения
EXPOSE 8080

#Устанавливаем рабочую директорию для запуска проекта
WORKDIR $SERVICE_HOME

# Запускаем сервис
CMD [ "./entrypoint.sh" ]