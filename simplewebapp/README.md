# Create and Run Simple Web App on GKE
Hashimoto Du at DevSamurai

Instruction for build and run as below

## Build Docker image
Require: Docker running on the build enviroment
```
# build docker image with tag name
docker build -t ds-gke-simplewebapp:latest -f simplewebapp.Dockerfile .
```

## Push docker image to GCP Container Registry
```
# login then set working project
gcloud auth login
gcloud config set project [PROJECT_ID]

# Configured Docker to use gcloud as a credential
gcloud auth configure-docker


# Tag the local image with the registry name 
# docker tag [SOURCE_IMAGE] [HOSTNAME]/[PROJECT-ID]/[IMAGE]:[TAG]
docker images
eg: docker tag 1e2055780000 asia.gcr.io/ds-project/ds-gke-simplewebapp:latest

# check image tag
docker images

# Push Docker image to Container Registry
# docker push [HOSTNAME]/[PROJECT-ID]/[IMAGE]
docker push asia.gcr.io/ds-project/ds-gke-simplewebapp

```

# create cluster on gke
```
# tạo cluster trong GKE
gcloud container clusters create ds-gke-small-cluster \
	--project ds-project \
	--zone asia-northeast1-b \
	--machine-type n1-standard-1 \
	--num-nodes 1 \
	--enable-stackdriver-kubernetes
```

```
# get credentials để truy cập cluster
gcloud container clusters get-credentials --zone asia-northeast1-b ds-gke-small-cluster

# deploy web app image to GKE
kubectl apply -f simplewebapp.deployment.yaml

```