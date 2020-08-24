FROM golang:1.15.0-alpine as builder

ADD ./ /go/src/api

WORKDIR /go/src/api

RUN go build -o api

# Stage 2
FROM alpine:3.12.0

COPY --from=builder /go/src/api /usr/bin/

ENTRYPOINT ["/usr/bin/api"]