ARG GO_VERSION=1.27

FROM golang:${GO_VERSION}-alpine AS base
WORKDIR /src
ENV CGO_ENABLED=0

FROM base AS deps
COPY go.mod go.sum ./
RUN go mod download

FROM deps AS builder
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/backend ./cmd/backend

FROM alpine:3.22 AS runner
ENV APP_ENV=production \
    HTTP_ADDR=:8082
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S -G app app
COPY --from=builder --chown=app:app /out/backend /usr/local/bin/backend
USER app
EXPOSE 8082
HEALTHCHECK --interval=10s --timeout=3s --start-period=20s --retries=5 \
  CMD wget -qO- http://127.0.0.1:8082/ping >/dev/null || exit 1
CMD ["backend"]
