FROM google/cloud-sdk:alpine
WORKDIR /app
COPY ./sample_app /app

ENTRYPOINT ["./sample_app"]