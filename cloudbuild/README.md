# CloudbuildでDockerイメージビルドとContainer Registryに登録


*記事の目的*

Dockerイメージビルドして、Container Registryに登録するまでは複数ステップがあって、毎回各ステップ別で実施するとめんどくさい。
それを解決するため、Cloudbuildを使って複数ステップをまとめてビルドを行います。

Cloudbuildを使わない場合、GKEにデプロイする手順

-   Dockerイメージのビルド
-   GCP Container RegistryにDockerイメージを登録（複数操作実施）
-   Container RegistryからアプリケーションをGKEにデプロイ

この記事では、Cloudbuildを使って最初の２つの手順をまとめて１回で実施できる。

-   Cloudbuildでイメージ作成とContainer Registryに登録（まとめて１回実施）
-   Container RegistryからアプリケーションをGKEにデプロイ


## 1.　サンプルアプリケーション準備

※以前記事のwebアプリケーションを再利用してCloudbuildを入れます。
[GKE上にwebアプリケーションを構築する方法](https://qiita.com/devs_hd/items/8edf3452d9912c19c7d8)

フォルダ構成

```sh
cloudbuild
├── README.md
├── cloudbuild.simplewebapp.yaml
├── go.mod
├── go.sum
├── simplewebapp.Dockerfile
├── simplewebapp.deployment.yaml
└── webapp
    ├── handler
    │   └── simplewebapp_handler.go
    └── simplewebapp.go
```


## 2.　Cloudbuildでイメージ作成とContainer Registryに登録

```sh:cloudbuild.simplewebapp.yaml
options:
  env:
  - GO111MODULE=on
  volumes:
  - name: go-modules
    path: /go

steps:
# go test
- name: golang:1.12
  dir: .
  args: ['go', 'test', './...']

# go build
- name: golang:1.12
  dir: .
  args: ['go', 'build', '-o', 'simplewebapp', 'webapp/simplewebapp.go']
  env: ["CGO_ENABLED=0"]

# docker build
- name: 'gcr.io/cloud-builders/docker'
  dir: .
  args: [
         'build',
         '-t', '${_GCR_REGION}/${_GCR_PROJECT}/${_GCR_IMAGE_NAME}:${_GCR_TAG}',
         '-f', 'simplewebapp.Dockerfile',
         '--cache-from', '${_GCR_REGION}/${_GCR_PROJECT}/${_GCR_IMAGE_NAME}:${_GCR_TAG}',
         '.'
        ]

# push image to Container Registry
- name: 'gcr.io/cloud-builders/docker'
  args: ["push", '${_GCR_REGION}/${_GCR_PROJECT}/${_GCR_IMAGE_NAME}']

substitutions:
  # # GCR region name to push image
  _GCR_REGION: asia.gcr.io
  # # Project ID
  _GCR_PROJECT: ds-project
  # # Image name
  _GCR_IMAGE_NAME: ds-cloudbuild-test
  # # Image tag
  _GCR_TAG: latest
```


```sh:simplewebapp.Dockerfile
FROM alpine:latest
WORKDIR /app
COPY ./simplewebapp /app

EXPOSE 80
ENTRYPOINT ["./simplewebapp"]
```

Cloudbuild実施

```sh:
cd path_to_app_folder

gcloud builds submit --config cloudbuild.simplewebapp.yaml
```
<img width="780" alt="gcp_gke_kubernetes_cloudbuild_dss_001.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/ad727737-d0ee-2b95-9b14-38cb6fe6a15d.png">


実施後、Container Registry上の作成イメージを確認
<img width="773" alt="gcp_gke_kubernetes_cloudbuild_dss_002.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/36109757-26ea-7ffd-ed10-5ef7276e15f6.png">



## 3.　Container RegistryからアプリケーションをGKEにデプロイ

デプロイメント定義ファイルを準備する。

```sh:simplewebapp.deployment.yaml
apiVersion: v1
kind: Service
metadata:
  name: simple-webapp-service
spec:
  ports:
  - port: 80
    targetPort: 80
    protocol: TCP
    name: http
  selector:
    app: simple-webapp
  type: LoadBalancer
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: simple-webapp
spec:
  replicas: 2
  template:
    metadata:
      labels:
        app: simple-webapp
    spec:
      containers:
      - name: simple-webapp
        image: asia.gcr.io/ds-project/ds-cloudbuild-test:latest
        ports:
          - containerPort: 80
```

このデプロイメント定義ファイルは、１つのLoad
Balancing作成、ポート80で公開する。webアプリケーションは後ろの２つのContainersで稼働とする。



```sh:
# k8sコントロールツールをインストール

gcloud components install kubectl
kubectl version

# GKEのクラスタにアクセスするため、credentialsを設定
gcloud container clusters get-credentials --zone asia-northeast1-b ds-gke-small-cluster

# GKEにアプリケーションを設定する
kubectl apply -f simplewebapp.deployment.yaml
```

デプロイ後、結果確認
<img width="926" alt="gcp_gke_kubernetes_cloudbuild_dss_003.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/23db8cab-0413-a296-621b-91c87112120c.png">

<img width="1138" alt="gcp_gke_kubernetes_cloudbuild_dss_004.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/54822e01-d399-5664-f31c-1a764ba0e01c.png">


外部からwebアプリケーションへのアクセス確認

```sh:
curl http://34.85.10.96/
> Hello, World!
```


<br>  
本記事で利用したソースコードはこちら
[https://github.com/dssolutioninc/dss_gke/tree/master/cloudbuild](https://github.com/dssolutioninc/dss_gke/tree/master/cloudbuild)

 
<br> 
最後まで読んで頂き、どうも有難う御座います!
DSS 橋本