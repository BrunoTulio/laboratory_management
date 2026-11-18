# syntax=docker/dockerfile:1

ARG GO_VERSION=1.27.1
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
WORKDIR /src

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download

ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 GOARCH=$TARGETARCH go build -ldflags="-s -w" -o /bin/server ./cmd/server

FROM alpine:3.21 AS final

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

ENV TZ=America/Bahia \
    APP_ENV=production \
    APP_PORT=8080 \
    DATABASE_PATH=/app/data/laboratory.db

ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser

WORKDIR /app

COPY --from=build /bin/server /app/server
COPY templates/ /app/templates/
COPY static/ /app/static/
COPY migrations/ /app/migrations/

RUN mkdir -p /app/data && chown -R appuser:appuser /app/data

USER appuser

VOLUME ["/app/data"]
EXPOSE 8080

ENTRYPOINT ["/app/server"]
