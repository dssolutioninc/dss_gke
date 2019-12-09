# GKE上スケジュールされるジョブを稼働させる方法


別の記事で[GKE上にバッチジョブを稼働させる方法](https://qiita.com/devs_hd/items/4679ad5075a513679766)を紹介しましたが、今回稼働スケジュールも設定してジョブデプロイする方法を紹介したいと思っております。

早速ですが、実施手順のまとめは下記となります。

- バッチジョブ作成
- イメージをビルドして、Container Registryにプッシュ
- ジョブ稼働スケジュールを設定し、デプロイする


※本手順はGCPとgcloudコマンドの利用経験があることが望ましいです。

ワークフォルダの構成

```sh
cron_job
├── build
│   ├── cloudbuild.cronjob.yaml
│   └── simplejob.Dockerfile
├── deployment
│   └── cronjob.deployment.yaml
├── go.mod
├── go.sum
└── job
    ├── handler
    │   └── simplejob_handler.go
    └── simplejob.go
```

## 1.　バッチジョブ作成
go言語を使って、簡単なバッチジョブを作成します。

```go:simplejob.go
package main

import (
	"context"
	"log"

	"github.com/itdevsamurai/gke/cronjob/job/handler"
)

func main() {
	log.Println("Job Started.")

	ctx := context.Background()
	handler.SimpleJobHandler{}.Run(ctx)

	log.Println("Job Finished.")
}
```

```go:simplejob_handler.go
package handler

import (
	"context"
	"flag"
	"log"
	"os/exec"
)

type SimpleJobHandler struct {
}

func (j SimpleJobHandler) Run(ctx context.Context) error {
	log.Println("Processing ...")

	runTime := getArguments()
	cmd := exec.Command("sleep", runTime)
	err := cmd.Run()
	if err != nil {
		log.Printf("Error at command: %v", cmd)
		return err
	}

	log.Println("Process completed.")
	return nil
}

func getArguments() string {
	runTime := flag.String("run-time", "", "specify number of seconds to run job")
	flag.Parse()
	return *runTime
}
```

このバッチジョブはインプットパラメータで稼働時間（何秒か）を指定する。

ローカルで稼働させてみる

```
cd path_to_cronjob
go run job/simplejob.go --run-time=5
```
<img width="553" alt="gcp_gke_kubernetes_batch_job_devsamurai_001.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/aa32f0ca-42b0-63e3-b867-2e00fc032b7c.png">



## 2.　イメージをビルドして、Container Registryにプッシュ

Cloud buildを使ってDockerイメージのビルドとContainer Registryにプッシュを行います。

```sh:simplejob.Dockerfile
FROM golang:1.12 as build_env

WORKDIR /go/src/github.com/itdevsamurai/gke/cronjob

COPY ./job ./job
COPY go.mod ./

ENV PATH="${PATH}:$GOPATH/bin"
ENV GO111MODULE=on

RUN export GOPROXY="https://proxy.golang.org" && export GO111MODULE=on && CGO_ENABLED=0 go build -o simplejob job/simplejob.go

FROM alpine:latest
WORKDIR /app
COPY --from=build_env /go/src/github.com/itdevsamurai/gke/cronjob /job
```

```sh:cloudbuild.cronjob.yaml
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
  args: ['go', 'build', '-o', 'simplejob', 'job/simplejob.go']
  env: ["CGO_ENABLED=0"]

# docker build
- name: 'gcr.io/cloud-builders/docker'
  dir: .
  args: [
         'build',
         '-t', '${_GCR_REGION}/${_GCR_PROJECT}/${_GCR_IMAGE_NAME}:${_GCR_TAG}',
         '-f', 'build/simplejob.Dockerfile',
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
  _GCR_PROJECT: project-abc123
  # # Image name
  _GCR_IMAGE_NAME: simple-job
  # # Image tag
  _GCR_TAG: latest
```

イメージビルドとプッシュ実行

```sh
cd cronjob_folder

gcloud builds submit --config build/cloudbuild.cronjob.yaml
```

実施後確認
<img width="909" alt="gcp_gke_kubernetes_cron_job_001.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/42377669-7a0a-2764-3724-4eb6db5be46a.png">



## 3.  ジョブ稼働スケジュールを設定し、デプロイする

デプロイメント定義ファイル作成

```sh:cronjob.deployment.yaml
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: simple-cron-job
spec:
  schedule: "*/5 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: simple-cron-job-container
            image: asia.gcr.io/project-abc123/simple-job:latest
            command: ["/job/simplejob"]
            args: ["--run-time", "10"]
          restartPolicy: Never
```

稼働スケジュールの設定は spec.schedule フィールドで指定します。
spec.schedule フィールドは、UNIX 標準の crontab 形式を使用して CronJob を実行する時間と間隔を定義します。すべての CronJob 時間は UTC で表示されます。スペースで区切られた 5 つのフィールドがあります。これらのフィールドは、以下を表します。

1. 分（0～59）
2. 時間（0～23）
3. 日（1～31）
4. 月（1～12）
5. 曜日（0～6）
すべての spec.schedule フィールドで、次の特殊文字を使用できます。

? は、単一の文字と一致するワイルドカード値です。
* は、ゼロ個以上の文字と一致するワイルドカード値です。
/ を使用すると、フィールドの間隔を指定できます。たとえば、最初のフィールド（分フィールド）の値が */5 の場合、「5 分ごと」を意味します。5 番目のフィールド（曜日フィールド）が 0/5 に設定されている場合、「5 回目の日曜日ごと」を意味します。

最後はジョブデプロイを実施します。

```
cd cronjob_folder

# スケジュールされたジョブをデプロイする
kubectl apply -f deployment/cronjob.deployment.yaml
```

デプロイ後の確認
<img width="945" alt="gcp_gke_kubernetes_cron_job_002.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/9dd67c8b-9df2-bc29-c4c4-cbfb8e95e59c.png">


これまでスケジュール稼働ジョブデプロイは完了となります。


本記事で利用したソースコードはこちら

[https://github.com/itdevsamurai/gke/tree/master/cronjob](https://github.com/itdevsamurai/gke/tree/master/cronjob)

 

最後まで読んで頂き、どうも有難う御座います!

DevSamurai 橋本