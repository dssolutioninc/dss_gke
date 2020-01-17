FROM alpine:latest
WORKDIR /app
COPY ./simplewebapp /app

EXPOSE 80
ENTRYPOINT ["./simplewebapp"]