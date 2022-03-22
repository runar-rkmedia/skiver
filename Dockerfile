FROM scratch

WORKDIR /app

# Go-releaser uses this path
COPY  ./skiver-api   ./skiver
# This is the _real_ path if building manually
# COPY  ./dist/skiver-api_linux_amd64/skiver-api ./skiver
# COPY ./dist/skiver-linux-amd64 ./skiver
ENTRYPOINT [ "/app/skiver" ]
