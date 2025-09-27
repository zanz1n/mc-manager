FROM golang:1 AS builder

ARG VERSION

WORKDIR /build

ENV DEBUG=0
ENV CGO_ENABLED=0
ENV SKIPTESTS=1
ENV OUTPUT=bin/node

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY Makefile .

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make deps

COPY . .

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make build-node

FROM gcr.io/distroless/static-debian12

COPY --from=builder /build/bin/node /node

ENTRYPOINT [ "/node" ]
