FROM golang:1.15.5-alpine-3.12 AS builder
WORKDIR /src/
ADD . /src/
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags='-extldflags=-static' -o /bot 

FROM gcr.io/distroless/static
COPY --from=builder /bot /bin/bot

ENTRYPOINT [ "/bin/bot" ]