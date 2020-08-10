FROM alpine:latest
WORKDIR /app
COPY ./task-service /app

EXPOSE 80
ENTRYPOINT ["./task-service"]