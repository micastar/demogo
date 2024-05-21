FROM alpine:latest

COPY ./bin/main /fss

ENTRYPOINT ["/fss", "i"]