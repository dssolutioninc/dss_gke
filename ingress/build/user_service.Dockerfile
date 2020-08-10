FROM alpine:latest
WORKDIR /app
COPY ./user-service /app

EXPOSE 80
ENTRYPOINT ["./user-service"]