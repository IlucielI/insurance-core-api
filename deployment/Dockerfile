FROM insurance-core-api-base:latest AS builder

ARG APP_VERSION=0.1.0
ARG GIT_HASH=dev

WORKDIR /src

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X github.com/bayuanugerah/insurance-core-api/internal/config.Version=${APP_VERSION} -X github.com/bayuanugerah/insurance-core-api/internal/config.GitHash=${GIT_HASH}" \
    -o /out/api ./cmd/api

FROM alpine:3.20 AS runtime

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S app -G app

COPY --from=builder /out/api /app/api

USER app

EXPOSE 8080

CMD ["/app/api"]
