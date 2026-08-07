# Етап 1: Збірка
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Копіюємо залежності
COPY go.mod go.sum ./
RUN go mod download

# Копіюємо весь код
COPY . .

# Збираємо бінарний файл
RUN CGO_ENABLED=0 GOOS=linux go build -o witness-api .

# Етап 2: Фінальний маленький образ
FROM alpine:3.20

RUN apk --no-cache add ca-certificates=20240705-r0

WORKDIR /app

# Копіюємо тільки зібраний бінарник
COPY --from=builder /app/witness-api .

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=3s CMD wget -q --spider http://localhost:8081/health || exit 1

CMD ["./witness-api"]