FROM golang:alpine as builder
WORKDIR /app
ARG ldflags
RUN echo CGO_ENABLED=0 GOOS=linux go build -ldflags="${ldflags}"  -ldflags="-s -w" -o skiver-api .
RUN apk update && apk upgrade && apk add --no-cache ca-certificates
RUN update-ca-certificates
COPY . .
# go1.18beta2 build -ldflags="${ldflags}" -o dist/skiver${SUFFIX} main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="${ldflags}" -o skiver-api .

FROM scratch

WORKDIR /app

# Go-releaser uses this path
# COPY  ./skiver-api   ./skiver-api
# This is the _real_ path if building manually
# COPY  ./dist/skiver-api_linux_amd64/skiver-api ./skiver
# COPY ./dist/skiver-linux-amd64 ./skiver
COPY --from=builder /app/skiver-api .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT [ "/app/skiver-api" ]
