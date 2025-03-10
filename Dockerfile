# Базовый образ
FROM golang:1.24

LABEL authors="Gleb Loginov and Nikita Kaprusov Ltd."

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -v -o /usr/local/bin/app ./cmd/auth

# Открываем порт
EXPOSE 8080

# Команда запуска
CMD ["/usr/local/bin/app"]