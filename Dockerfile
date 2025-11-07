FROM golang:1.25.1-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify
COPY cmd/app/ ./cmd/app
COPY internal/app/ ./internal/app
COPY data ./data

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app cmd/app/main.go

FROM alpine AS runner

RUN apk add --no-cache tzdata

RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser

WORKDIR /app

COPY --from=builder /app/data ./data
COPY --from=builder /app/app .

RUN chown -R appuser:appgroup /app

USER appuser

CMD ["./app"]