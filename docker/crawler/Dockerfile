FROM golang:1.18-alpine3.14 as builder

WORKDIR /app

RUN apk add --no-cache make git build-base

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN make build

FROM builder as tester

RUN apk add --no-cache g++ chromium

CMD make test

FROM alpine:3.15 as client
WORKDIR /app

COPY --from=builder /app/web-crawler-client .

CMD /app/web-crawler-client

FROM alpine:3.15 as api
WORKDIR /app

COPY --from=builder /app/web-crawler-api .

CMD /app/web-crawler-api

FROM alpine:3.15 as worker
WORKDIR /app

RUN apk add --no-cache chromium

COPY --from=builder /app/web-crawler-worker .

CMD /app/web-crawler-worker
