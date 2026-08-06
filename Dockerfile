# Етап 1: Збірка
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Копіюємо залежності
COPY go.mod go.sum ./
RUN go mod download

# Копіюємо весь код
COPY . .

# Збираємо бінарний файл
RUN CGO_ENABLED=0 GOOS=linux go build -o todo-api .

# Етап 2: Фінальний маленький образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копіюємо тільки зібраний бінарник
COPY --from=builder /app/todo-api .

EXPOSE 8081

CMD ["./todo-api"]