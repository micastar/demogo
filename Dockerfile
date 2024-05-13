FROM --platform=linux/amd64 golang:alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o feed ./cmd/main.go

CMD ["./feed"]