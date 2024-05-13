FROM --platform=$BUILDPLATFORM golang:alpine AS builder

ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY ./ ./

RUN CGO_ENABLED=0 GOARCH=$TARGETARCH GOOS=linux go build -a -o feed ./cmd/main.go


FROM alpine:latest

WORKDIR /root

COPY --from=builder /app/feed .

CMD /root/feed
