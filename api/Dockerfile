FROM golang:alpine as builder

RUN mkdir -p /build

ADD . /build/

WORKDIR /build

RUN go build -o main .

# stage 2

FROM alpine

RUN mkdir -p /app

RUN adduser -S -D -H -h /app appuser

USER appuser

WORKDIR /app

COPY --from=builder /build/main /app/

COPY --from=builder /build/.env /app/

EXPOSE 3000

CMD ["main"]