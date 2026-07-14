# syntax=docker/dockerfile:1.7
FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY cmd/shortlink-demo ./cmd/shortlink-demo
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_TIME=unknown
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o /out/shortlink-api ./cmd/shortlink-demo

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/shortlink-api /shortlink-api
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/shortlink-api"]
