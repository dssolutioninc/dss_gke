# Use Cloud Build to build docker image and push it on Container Registry
Hashimoto Du at DevSamurai

Instruction for build and run as below

## Run Cloud Build
Require: gcloud auth login, and set project first
```
cd path_to_app_folder
gcloud builds submit --config cloudbuild.simplewebapp.yaml
```

## Deploy app from Container registry image to GKE
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