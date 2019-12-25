# GKE中のGolangアプリケーションからCloud Pub/Subを使ってデータ連携を行う

GKEの中に稼働されるアプリケーションからどうやってGCPサービスにアクセスしたり、データ連携したりするか？という疑問がある方々に回答する記事をまとめました。
Pub/Subサービスを使って、サンプルとして作成しました。

手順まとめ

1. サービスアカウント作成＆アクセス用のaccount.jsonファイル発行
- Pub/SubのTopic＆Subscription作成
- Pub/SubをSubscriptionアプリケーションの準備
- ローカルで稼働確認
- Cloudbuildでアプリケーションのイメージビルド
- GKEのCluster準備
- GKEにアプリケーションをデプロイ
- 稼働検証

フォルダ構成

```sh
accesspubsub
├── README.md
├── app
│   ├── handler
│   │   └── sample_handler.go
│   └── sample_app.go
├── cloudbuild.sampleapp.yaml
├── deployment.sampleapp.yaml
├── go.mod
├── go.sum
├── sampleapp.Dockerfile
├── secret
│   └── account.json
└── tool
    └── publish_to_topic.go
```

## 1.　サービスアカウント作成＆アクセス用のaccount.jsonファイル発行

#### 　サービスアカウント作成
```sh
# set work project
gcloud config set project [PROJECT_ID]

# create service account
gcloud iam service-accounts create service-account \
  --display-name "Account using to call GCP service"
```

#### 　account.jsonファイル発行
```sh
# create service account's credential file
gcloud iam service-accounts keys create {{path_to_save/account.json}} \
  --iam-account service-account@[PROJECT_ID].iam.gserviceaccount.com
```

#### 　サービスアカウント権限付与
本記事は簡単とするため、editorロールを付与します。

```sh
# ロールをサービスアカウトに付与。下記のコマンドを実施するため、オーナー権限必要
# editor権限付与
gcloud projects add-iam-policy-binding [PROJECT_ID] \
  --member serviceAccount:service-account@[PROJECT_ID].iam.gserviceaccount.com \
  --role roles/editor
```

サービスアカウント権限付与について、他の付与方法はこの記事をご参考
[GCPのサービスを利用権限のまとめ](https://qiita.com/bendevs/items/066194b37e3753d0c201)

## 2.　Pub/SubのTopic＆Subscription作成
データ連携用のPub/SubのTopicとSubscriptionを作成する。

```sh
# Topic作成
gcloud pubsub topics create [TOPIC_NAME]
# 例
gcloud pubsub topics create sample-app-topic

# Topicのsubscription作成
gcloud pubsub subscriptions create [SUB_NAME] --topic=[TOPIC_NAME]
# 例
gcloud pubsub topics create sample-app-topic-sub
```


## 3.　Pub/SubをSubscriptionアプリケーションの準備

```go:sample_app.go
package main

import (
	"log"
	"os"

	"github.com/itdevsamurai/gke/accesspubsub/app/handler"
)

func main() {
	log.Println("Application Started.")

	// projectID is identifier of project
	projectID := os.Getenv("PROJECT_ID")

	// pubsubSubscriptionName use to hear the comming request
	pubsubSubscriptionName := os.Getenv("PUBSUB_SUBSCRIPTION_NAME")

	err := handler.SampleHandler{}.StartWaitMessageOn(projectID, pubsubSubscriptionName)
	if err != nil {
		log.Println("Got Error.")
	}

	log.Println("Application Finished.")
}
```

```go:sample_handler.go
package handler

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
)

type SampleHandler struct {
}

// StartWaitMessageOn
// projectID := "my-project-id"
// subName := projectID + "-example-sub"
func (h SampleHandler) StartWaitMessageOn(projectID, subName string) error {
	log.Println(fmt.Sprintf("StartWaitMessageOn [Project: %s, Subscription Name: %s]", projectID, subName))

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	sub := client.Subscription(subName)
	err = sub.Receive(ctx, processMessage)
	if err != nil {
		return err
	}

	return nil
}

// processMessage implement callback function to process received message data
var processMessage = func(ctx context.Context, m *pubsub.Message) {
	log.Println(fmt.Sprintf("Message ID: %s\n", m.ID))
	log.Println(fmt.Sprintf("Message Time: %s\n", m.PublishTime.String()))

	log.Println(fmt.Sprintf("Message Attributes:\n %v\n", m.Attributes))

	log.Println(fmt.Sprintf("Message Data:\n %s\n", m.Data))

	m.Ack()
}
```

このアプリケーションはPub/Subの指定Subscriptionをヒアリングして、メッセージがきたら、処理を行います。
処理はメッセージの内容を印刷するだけのシンプル処理となります。


## 4.　ローカルで稼働確認
稼働を検証するため、Pub/SubにメッセージをPublishするツールを準備します。

```sh:tool/publish_to_topic.go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
)

var (
	topicID   = flag.String("topic-id", "sample-topic", "Specify topic to publish message")
	projectID = flag.String("project-id", "sample-project", "Specify GCP project you want to work on")
)

func main() {
	flag.Parse()

	err := publishMsg(*projectID, *topicID,
		map[string]string{
			"user":    "Hashimoto",
			"message": "more than happy",
			"status":  "bonus day!",
		},
		nil)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
}

func publishMsg(projectID, topicID string, attr map[string]string, msg map[string]string) error {
	// projectID := "my-project-id"
	// topicID := "my-topic"
	// msg := message data publish to topic
	// attr := attribute of message
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}

	bMsg, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("Input msg error : %v", err)
	}

	t := client.Topic(topicID)
	result := t.Publish(ctx, &pubsub.Message{
		Data:       bMsg,
		Attributes: attr,
	})

	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("Get: %v", err)
	}
	fmt.Printf("Published message with custom attributes; msg ID: %v\n", id)

	return nil
}
```

ローカルで稼働検証

```sh
# GCPサービスにアクセスためのアカウントJSONファイルを環境変数に設定
# Windowsを使う方は環境変数の設定画面から行ってください。
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/account.json"

# アップリケーションを実行
export PROJECT_ID="project-abc123" && \
export PUBSUB_SUBSCRIPTION_NAME="sample-app-topic" && \
go run ./app/sample_app.go

# 別のターミナルを開いて、テストツールを実行
# テストのメッセージをTopicにPublishする
go run tool/publish_to_topic.go --project-id=project-abc123 --topic-id=sample-app-topic
```

ローカル稼働のコンソールログ
<img width="1025" alt="gke_kubernetes_pubsub_access_001.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/dc70f96b-c829-fe06-ba6b-e34d5f5269b9.png">


## 5.　Cloudbuildでアプリケーションのイメージビルド

Dockerfile作成

```sh:sampleapp.Dockerfile
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY ./sample_app /app

ENTRYPOINT ["./sample_app"]
```

Pub/Subにアクセスするため、サービスアカウントのJSONファイルで認証します。
「alpine」イメージのみはライブラリーが足りなくて、認証仕組みはエラーとなります。
「ca-certificates」ライブラリーを追加する必要です。これは注意点となります。

```sh:cloudbuild.sampleapp.yaml
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
  args: ['go', 'build', '-o', 'sample_app', 'app/sample_app.go']
  env: ["CGO_ENABLED=0"]

# docker build
- name: 'gcr.io/cloud-builders/docker'
  dir: .
  args: [
         'build',
         '-t', '${_GCR_REGION}/${_GCR_PROJECT}/${_GCR_IMAGE_NAME}:${_GCR_TAG}',
         '-f', 'sampleapp.Dockerfile',
         '--cache-from', '${_GCR_REGION}/${_GCR_PROJECT}/${_GCR_IMAGE_NAME}:${_GCR_TAG}',
         '.'
        ]

# push image to Container Registry
- name: 'gcr.io/cloud-builders/docker'
  args: ["push", '${_GCR_REGION}/${_GCR_PROJECT}/${_GCR_IMAGE_NAME}']

substitutions:
  # # Project ID
  _GCR_PROJECT: project-abc123
  # # GCR region name to push image
  _GCR_REGION: asia.gcr.io
   # # Image name
  _GCR_IMAGE_NAME: sample-pubsub-usage-app
  # # Image tag
  _GCR_TAG: latest
```

アプリケーションのイメージビルド。

```sh
gcloud builds submit --config cloudbuild.sampleapp.yaml
```

ビルド完了となったらContainer Registryで確認する。
<img width="776" alt="gke_kubernetes_pubsub_access_002.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/86ea1662-a3c5-e4a9-4705-2fdb052539d2.png">



Cloudbuildについてさらに確認したい場合、この記事にご参考（[CloudbuildでDockerイメージビルドとContainer Registryに登録](https://qiita.com/devs_hd/items/04a71fb764a1ae492bb5)）


## 6.　GKEのCluster準備

```sh
# クラスタ作成
gcloud container clusters create ds-gke-small-cluster \
	--project ds-project \
	--zone asia-northeast1-b \
	--machine-type n1-standard-1 \
	--num-nodes 1 \
	--enable-stackdriver-kubernetes

# k8sコントロールツールをインストール
gcloud components install kubectl
kubectl version

# GKEのクラスタにアクセスするため、credentialsを設定
gcloud container clusters get-credentials --zone asia-northeast1-b ds-gke-small-cluster
```

## 7.　GKEにアプリケーションをデプロイ

アカウントJSONファイルを「secret generic」ボリュームとしてクラスタに登録します。

```
kubectl create secret generic service-account-credential \
     --from-file=./secret/account.json
```
<img width="1130" alt="gke_kubernetes_pubsub_access_003.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/b4522138-e837-56a1-6d74-8e38a12d6f4c.png">


デプロイ定義ファイルの準備

```sh:deployment.sampleapp.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: sample-pubsub-usage-app
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: sample-pubsub-usage-app
    spec:
      volumes:
      - name: service-account-credential
        secret:
          secretName: service-account-credential
      containers:
      - name: sample-pubsub-usage-app-container
        image: asia.gcr.io/project-abc123/sample-pubsub-usage-app:latest
        # environment variables for the Pod
        env:
        - name: PROJECT_ID
          value: "project-abc123"
        - name: PUBSUB_SUBSCRIPTION_NAME
          value: "sample-app-topic-sub"
        - name: GOOGLE_APPLICATION_CREDENTIALS
          value: /app/secret/account.json
        
        volumeMounts:
        - mountPath: /app/secret
          name: service-account-credential
          readOnly: true
```
【デプロイ定義の説明】
作成できた「secret generic」ボリュームをデプロイ定義にMount設定して、account.jsonファイルパスを環境変数に渡す。
「GOOGLE_APPLICATION_CREDENTIALS」の環境変数はPub/Subにアクセスするためプログラムは使います。この設定はポイントとなります。


アプリケーションをGKEにデプロイします。

```sh
kubectl apply -f deployment.sampleapp.yaml
```

GKE上のデプロイできたアプリケーションを確認
<img width="989" alt="gke_kubernetes_pubsub_access_004.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/d5ed9bdb-137b-d240-59b8-6b9380b4aec5.png">


## 8.　稼働検証

TopicにメッセージのPublishを行います。

```sh:
go run tools/publish_to_topic.go --project-id=project-abc123 --topic-id=sample-app-topic
```

GKE中のアプリケーションの稼働ログを確認
<img width="1385" alt="gke_kubernetes_pubsub_access_005.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/2434ad0b-d995-ba27-7658-31e75ccf1312.png">


<br>  
本記事で利用したソースコードはこちら
[https://github.com/itdevsamurai/gke/tree/master/accesspubsub](https://github.com/itdevsamurai/gke/tree/master/accesspubsub)

 
<br> 
最後まで読んで頂き、どうも有難う御座います!
DevSamurai 橋本
