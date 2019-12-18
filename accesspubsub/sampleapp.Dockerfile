FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY ./sample_app /app

ENTRYPOINT ["./sample_app"]