FROM golang:1.17.1-alpine3.14 AS builder
WORKDIR /src/
COPY . /src/
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags='-extldflags=-static' -o /bot 

FROM gcr.io/distroless/static
COPY --from=builder /bot /bin/bot

ENTRYPOINT [ "/bin/bot" ]