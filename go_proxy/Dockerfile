FROM golang:1.19 AS builder

ENV GO111MODULE=on

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/main /main

CMD ["/main"]
