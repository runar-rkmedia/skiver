FROM golang:alpine as builder
WORKDIR /app
ARG ldflags
RUN echo CGO_ENABLED=0 GOOS=linux go build -ldflags="${ldflags}"  -ldflags="-s -w" -o skiver-api .
RUN apk update && apk upgrade && apk add --no-cache ca-certificates
RUN update-ca-certificates
COPY . .
# go1.18beta2 build -ldflags="${ldflags}" -o dist/skiver${SUFFIX} main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="${ldflags}" -o skiver-api .

FROM alpine as alpine

WORKDIR /app

COPY --from=builder /app/skiver-api .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT [ "/app/skiver-api" ]

FROM scratch as scratch

WORKDIR /app

COPY --from=builder /app/skiver-api .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT [ "/app/skiver-api" ]
