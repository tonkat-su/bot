FROM golang:1.20-alpine3.17 AS builder
WORKDIR /src/
COPY . /src/
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags='-extldflags=-static' -o /interactions-server ./cmd/interactions-server/main.go

FROM gcr.io/distroless/static
COPY --from=builder /interactions-server /bin/interactions-server
EXPOSE 8080

ENTRYPOINT [ "/bin/interactions-server" ]
