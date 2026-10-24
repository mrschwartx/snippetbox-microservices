# syntax=docker/dockerfile:1
ARG GO_VERSION=1.27

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
COPY . .
RUN go mod tidy && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/backend ./cmd/backend

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S -G app app
COPY --from=build /out/backend /usr/local/bin/backend
USER app
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=20s --retries=5 \
  CMD wget -qO- http://127.0.0.1:8081/ping >/dev/null || exit 1
ENTRYPOINT ["backend"]
