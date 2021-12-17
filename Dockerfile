FROM gcr.io/distroless/base

WORKDIR /app
COPY  ./dist/skiver-linux-amd64   ./skiver
CMD [ "/app/skiver" ]
