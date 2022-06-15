FROM golang:1.18-alpine3.16 AS builder
WORKDIR /src/
COPY . /src/
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags='-extldflags=-static' -o /bot 

FROM gcr.io/distroless/static
COPY --from=builder /bot /bin/bot

ENTRYPOINT [ "/bin/bot" ]