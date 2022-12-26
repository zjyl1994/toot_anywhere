FROM golang:1.18-alpine AS builder
COPY . /src
WORKDIR /src
RUN go build -ldflags="-s -w" -o /app .
FROM alpine:3.17
RUN apk add --no-cache tini
COPY --from=builder /app /app/tootanywhere
COPY --from=builder /src/web /app/web
WORKDIR /app
VOLUME /app/data
EXPOSE 9900
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/tootanywhere"]