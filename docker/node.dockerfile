FROM docker.io/library/golang:1 AS builder

ARG VERSION

WORKDIR /build

ENV DEBUG=0
ENV CGO_ENABLED=0
ENV SKIPTESTS=1
ENV OUTPUT=bin/node

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache
RUN go env -w GOBIN=/usr/bin

COPY Makefile .

RUN go install github.com/bufbuild/buf/cmd/buf@latest

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make deps

COPY . .

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make build-node

FROM gcr.io/distroless/static-debian13

COPY --from=builder /build/bin/node /node

ENTRYPOINT [ "/node" ]
