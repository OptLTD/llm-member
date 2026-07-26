# Multi-stage build

FROM node:20-alpine AS admin-builder
WORKDIR /app/ui-admin
COPY ui-admin/package*.json ./
RUN npm ci
COPY ui-admin/ ./
RUN npm run build

FROM golang:1.24-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev sqlite-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=admin-builder /app/ui-front/admin/dist ./ui-front/admin/dist
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata sqlite
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' > /etc/timezone
RUN addgroup -g 1001 -S appgroup && adduser -u 1001 -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/ui-front ./ui-front
COPY --from=builder /app/locales ./locales
COPY --from=builder /app/.env.example ./.env.example
RUN mkdir -p ./data && chown -R appuser:appgroup /app
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1
CMD ["./main"]
