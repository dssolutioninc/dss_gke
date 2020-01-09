FROM alpine:latest
WORKDIR /app
COPY ./sample_app /app

ENTRYPOINT ["./sample_app"]