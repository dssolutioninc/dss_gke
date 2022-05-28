# Cloud Endpointsを使って、GKEのサービス間の認証

## 　1．はじめに
別の記事でGKEにウェブサービスをデプロイする方法を紹介いたしました。
そのウェブサービスはどうやってアクセス元を認証するかのご質問がある方々がいると思います。
ウェブサービスに認証仕組みを追加する方法もありますが、手間がかかります。
本記事は GCP Cloud Endpoints を使うGKEのサービス間の認証方法を紹介いたします。


## 　2．アーキテクチャー
早速ですが、想像しやすくするため、全体アーキテクチャーを先に見せます。
![gke_service_authen_between_services_001.png](https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/54e3b133-e924-37a3-9c86-f6205e5ada1d.png)


これから構築に行きます。

## 　3．アプリケーションのフォルダ構成

```sh
application_folder
├── build
│   ├── cloudbuild.simplewebapp.yaml
│   └── simplewebapp.Dockerfile
├── deployment
│   └── simplewebapp.deployment.yaml
├── go.mod
├── go.sum
├── openapiv2
│   └── simple_webapp_endpoint.yaml
├── tools
│   ├── api
│   │   └── config
│   │       └── config.go
│   └── call_sercure_api.go
└── webapp
    ├── handler
    │   └── simplewebapp_handler.go
    └── simplewebapp.go
```

## 　4．ウェブアプリケーション
下記の３つURLを提供します。

```sh
/index    : Hello World を表示
/public   : Public を表示
/private  : Private を表示
```

詳細な処理はソースコードをご参照。


## 　5．Cloud Endpoints

サービス間認証の場合、サービスアカウントを使う認証方法がお進められます。

サービスアカウントとアカウントJSONファイルの作成

```sh
gcloud iam service-accounts create service_account_001 --display-name "service account for authentication between services on gke"
  
gcloud iam service-accounts list

# サービス アカウント トークン作成者の役割を追加します
gcloud projects add-iam-policy-binding PROJECT_ID \
  --member serviceAccount:SA_EMAIL_ADDRESS \
  --role roles/serviceAccountTokenCreator

# 現在の作業ディレクトリにサービス アカウント キー ファイルを作成します。FILE_NAME は、キーファイルに使用する名前に置き換えます。デフォルトでは、gcloud JSON ファイルが作成されます。
gcloud iam service-accounts keys create FILE_NAME.json \
  --iam-account SA_EMAIL_ADDRESS
```

Static IP作成

```sh
# Check static IP.
gcloud compute addresses list

# Create static IP.
gcloud compute addresses create simple-webapp-service-ip --region asia-northeast1

# Check static IP.
gcloud compute addresses list
> NAME                                           ADDRESS/RANGE   TYPE      PURPOSE  NETWORK  REGION           SUBNET  STATUS
> simple-webapp-service-ip                       35.243.84.30    EXTERNAL                    asia-northeast1          RESERVED
```
「35.243.84.30」のIPをメモして、下記のOpenAPI定義で使います。

Cloud Endpointsを作るため、OpenAPI定義を作成に行きます。


```yaml:simple_webapp_endpoint.yaml
# [START swagger]
swagger: "2.0"
info:
  description: "Simple Sercure Webapp API Endpoints"
  version: 1.0.0
  title: Simple Sercure Webapp
host: "simple-sercure-webapp.endpoints.PROJECT_ID.cloud.goog"
x-google-endpoints:
  - name: "simple-sercure-webapp.endpoints.PROJECT_ID.cloud.goog"
    target: "35.243.84.30"

tags:
  - name: simple-webapp
    description: Simple Sercure Webapp
# [END swagger]
consumes:
- "application/json"
- "text/html"
produces:
- "application/json"
- "text/html"
schemes:
- "http"
paths:
  "/index":
    get:
      description: "index page"
      operationId: "index"
      produces:
      - "text/html"
      responses:
        200:
          description: "Success"
          schema:
            type: array
  "/public":
    get:
      description: "public page"
      operationId: "public"
      produces:
      - "text/html"
      responses:
        200:
          description: "Success"
          schema:
            type: array
  "/private":
    get:
      description: "private page"
      operationId: "private"
      produces:
      - "text/html"
      responses:
        200:
          description: "Success"
          schema:
            type: array
      security:
      - google_jwt: []

securityDefinitions:
  # This section configures authentication using Google API Service Accounts
  # to sign a json web token. This is mostly used for server-to-server
  # communication.
  google_jwt:
    authorizationUrl: ""
    flow: "implicit"
    type: "oauth2"
    # This must match the 'iss' field in the JWT.
    x-google-issuer: "service_account_001@PROJECT_ID.iam.gserviceaccount.com"
    # Update this with your service account's email address.
    x-google-jwks_uri: "https://www.googleapis.com/robot/v1/metadata/x509/service_account_001@PROJECT_ID.iam.gserviceaccount.com"
```

PROJECT_IDはデプロイ先のProject IDを入れ替えてください。
「securityDefinitions」では認証方法を定義されます。今回はサービスアカウントを利用して認証させるとする。
認証対象のAPIは

```sh
  - google_jwt: []
```
追加します。今回 /private のAPIのみ認証対象となります。

Cloud Endpoints にデプロイ

```sh
# deploy endpoints
gcloud endpoints services deploy openapiv2/simple_webapp_endpoint.yaml
```

デプロイ後の確認
<img width="1418" alt="gke_service_authen_between_services_002.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/cc48329f-0e5b-bf1f-facc-c2480095db10.png">


## 　6．Container Registryにイメージビルド

```sh
gcloud builds submit --config build/cloudbuild.simplewebapp.yaml
```

## 　7．GKEにデプロイメントの定義

上記の構成図どおり、認証できるため「Extensible Service Proxy」を使います。
Extensible Service Proxy（ESP）は、OpenAPI または gRPC API バックエンドの前面で動作し、認証、モニタリング、ロギングなどの API 管理機能を提供する高パフォーマンスでスケーラブルなプロキシです。

```yaml:simplewebapp.deployment.yaml
apiVersion: v1
kind: Service
metadata:
  name: simple-webapp-service
spec:
  ports:
  - port: 80
    targetPort: 8081
    protocol: TCP
    name: http
  selector:
    app: simple-webapp
  type: LoadBalancer
  loadBalancerIP: "35.243.84.30"
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
      - name: esp
        image: gcr.io/endpoints-release/endpoints-runtime:1
        args: [
          "--http_port", "8081",
          "--backend", "127.0.0.1:8080",
          "--service", "simple-sercure-webapp.endpoints.PROJECT_ID.cloud.goog",
          "--rollout_strategy", "managed",
        ]
        ports:
          - containerPort: 8081
      - name: simple-webapp
        image: asia.gcr.io/PROJECT_ID/simple-webapp:latest
        ports:
          - containerPort: 8080
```

Endpoints のサービス名 「simple-sercure-webapp.endpoints.PROJECT_ID.cloud.goog」をESPのパラメタに指定する必要です。

GKEにデプロイ

```sh
kubectl apply -f deployment/simplewebapp.deployment.yaml

# デプロイ後のサービス情報確認
kubectl get services
> NAME                    TYPE           CLUSTER-IP     EXTERNAL-IP    PORT(S)        AGE
> kubernetes              ClusterIP      10.23.240.1    <none>         443/TCP        49d
> simple-webapp-service   LoadBalancer   10.23.253.95   35.243.84.30   80:30169/TCP   18m
```
EXTERNAL-IP は指定された Static IP となっています。

GKEで確認
<img width="1145" alt="gke_service_authen_between_services_003.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/69f217b3-8d1b-311c-f6b1-2425d01648e9.png">
<img width="1139" alt="gke_service_authen_between_services_004.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/14e576ac-8c88-5878-f274-d5eedc0a99d1.png">


## 　8．稼働確認
ブラウザでStatic IPを開いて確認してみます。
<img width="564" alt="gke_service_authen_between_services_005.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/8a2511e8-2dc4-b158-bfc2-ea2b4220b77c.png">

<img width="569" alt="gke_service_authen_between_services_006.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/843e067c-ff89-1a6d-6d53-b9bfc6f2b618.png">

PrivateのURLを認証エラーとなりました。サービスアカウントのJWTトークンを使ってアクセスしないといけない。

JWTトークンを生成して、Private API をコールするツールを準備しました。
実施する前に、サービスアカウント情報を設定します。

```go:tools/api/config/config.go
package config

var (
	Audience            = "simple-sercure-webapp.endpoints.PROJECT_ID.cloud.goog"
	ServiceAccountFile  = "path_to/service_account_001.json"
	ServiceAccountEmail = "service_account_001@PROJECT_ID.iam.gserviceaccount.com"

	ApiUrl        = "http://simple-sercure-webapp.endpoints.PROJECT_ID.cloud.goog/private"
	RequestMethod = "GET"
	RequestData   = map[string]interface{}{}
)
```

Go〜
<img width="658" alt="gke_service_authen_between_services_007.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/f9104fa9-948e-1d86-6873-7655e3074f2a.png">
JWTトークンを使って、Private URLをアクセス出来ました。

これで以上となります。

<br>  
本記事で利用したソースコードはこちら
[https://github.com/dssolutioninc/dss_gke/tree/master/sercure_k8s_service_sa](https://github.com/dssolutioninc/dss_gke/tree/master/sercure_k8s_service_sa)

 
<br> 
最後まで読んで頂き、どうも有難う御座います!
DSS 橋本