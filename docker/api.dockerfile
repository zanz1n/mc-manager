FROM docker.io/oven/bun:latest AS bun_builder

WORKDIR /app

COPY ./web/package.json ./web/bun.lock /app/
RUN bun install --frozen-lockfile

COPY ./web .
RUN bun run build

FROM docker.io/library/golang:1 AS builder

ARG VERSION

WORKDIR /build

ENV DEBUG=0
ENV CGO_ENABLED=0
ENV SKIPTESTS=1
ENV OUTPUT=bin/api

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache
RUN go env -w GOBIN=/usr/bin

COPY Makefile .

RUN go install github.com/bufbuild/buf/cmd/buf@latest

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make deps

COPY . .
COPY --from=bun_builder /app/build /build/web/build

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make build-api

FROM gcr.io/distroless/static-debian13

COPY --from=builder /build/bin/api /api

ENTRYPOINT [ "/api" ]
