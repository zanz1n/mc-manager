FROM oven/bun AS bun_builder

WORKDIR /app

COPY ./web/package.json ./web/bun.lock /app/
RUN bun install --frozen-lockfile

COPY ./web .
RUN bun run build

FROM golang:1 AS builder

ARG VERSION

WORKDIR /build

ENV DEBUG=0
ENV CGO_ENABLED=0
ENV SKIPTESTS=1
ENV OUTPUT=bin/api

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY Makefile .

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make deps

COPY . .
COPY --from=bun_builder /app/build /build/web/build

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make build-api

FROM gcr.io/distroless/static-debian12

COPY --from=builder /build/bin/api /api

ENTRYPOINT [ "/api" ]
