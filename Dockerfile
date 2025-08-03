# Build Go backend
FROM golang:1.24.5-alpine3.22 AS go-builder
WORKDIR /app
COPY . .
RUN go build -o main main.go
RUN apk add --no-cache curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate migrate.linux-amd64

# Build Astro frontend
FROM node:22-alpine AS web-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Production runtime
FROM alpine:3.22
WORKDIR /app
RUN apk add --no-cache netcat-openbsd

# Copy Go backend
COPY --from=go-builder /app/main .
COPY --from=go-builder /app/migrate.linux-amd64 ./migrate
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
RUN chmod +x start.sh
RUN chmod +x wait-for.sh

# Copy database migrations
COPY db/migration ./migration

# Copy built frontend
COPY --from=web-builder /app/web/dist ./web/dist

EXPOSE 8080
EXPOSE 4321
CMD ["/app/main"]
ENTRYPOINT [ "/app/start.sh" ]