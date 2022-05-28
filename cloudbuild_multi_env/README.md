# 複数環境に向けて、GCP Cloudbuildスクリプトの構成

開発・ステージング・本番などの複数環境に向けてGCP Cloudbuildの構成を紹介したいと思います。
シンプルのウェブアプリケーションを使ってビルドを行います。

フォルダ構成

```sh
cloudbuild_multi_env
├── README.md
├── build
│   ├── cloudbuild
│   │   ├── _base
│   │   │   └── cloudbuild.simplewebapp.yaml
│   │   ├── dev
│   │   │   └── cloudbuild.simplewebapp.yaml
│   │   ├── prod
│   │   │   └── cloudbuild.simplewebapp.yaml
│   │   └── staging
│   │       └── cloudbuild.simplewebapp.yaml
│   └── dockerfile
│       └── simplewebapp.Dockerfile
└── webapp
    ├── handler
    │   └── simplewebapp_handler.go
    └── simplewebapp.go
```

## 1.　サンプルウェブアプリケーション準備

過去の記事用のウェブアップリケーションを再利用します。
[GKE上にwebアプリケーションを構築する方法](https://qiita.com/devs_hd/items/8edf3452d9912c19c7d8)


```go:simplewebapp.go
package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/dssolutioninc/dss_gke/simplewebapp/webapp/handler"
)

// Default Server Port
const DEFAULT_SERVER_PORT = ":80"

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Route => handler
	e.GET("/", handler.SimpleWebHandler{}.Index)

	e.GET("/ping", handler.SimpleWebHandler{}.Ping)

	// Start server
	e.Logger.Fatal(e.Start(DEFAULT_SERVER_PORT))
}
```

```go:simplewebapp_handler.go
package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type SimpleWebHandler struct {
}

func (sh SimpleWebHandler) Index(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!\n")
}

func (sh SimpleWebHandler) Ping(c echo.Context) error {
	return c.String(http.StatusOK, "Pong!\n")
}
```

## 2.　Dockerfileの準備

```sh:simplewebapp.Dockerfile
FROM alpine:latest
WORKDIR /app
COPY ./simplewebapp /app

EXPOSE 80
ENTRYPOINT ["./simplewebapp"]
```

## 3.　Cloudbuildスクリプトの準備

### 各種環境の共有スクリプト

```sh:_base/cloudbuild.simplewebapp.yaml
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
         '-f', 'build/dockerfile/simplewebapp.Dockerfile',
         '--cache-from', '${_GCR_REGION}/${_GCR_PROJECT}/${_GCR_IMAGE_NAME}:${_GCR_TAG}',
         '.'
        ]

# push image to Container Registry
- name: 'gcr.io/cloud-builders/docker'
  args: ["push", '${_GCR_REGION}/${_GCR_PROJECT}/${_GCR_IMAGE_NAME}']

substitutions:
  # # GCR region name to push image
  _GCR_REGION: asia.gcr.io
  # # Image name
  _GCR_IMAGE_NAME: ds-cloudbuild-test
  # # Image tag
  _GCR_TAG: latest

```

### 各種環境別スクリプト

こつは「--substitutions」を使って環境別の設定を指定します。今回はプロジェクトIDだけとなりますが、他の設定があったらコンマで区切って追加ください。
例：

```sh:
'--substitutions=_GCR_PROJECT=ds-abc123-dev,_ENV=dev'
```

開発環境

```sh:dev/cloudbuild.simplewebapp.yaml
steps:
- name: 'gcr.io/cloud-builders/gcloud'
  args: [
      'builds', 
      'submit',
      '--config=build/cloudbuild/_base/cloudbuild.simplewebapp.yaml',
      '--substitutions=_GCR_PROJECT=ds-abc123-dev',
      '.'
  ]
```

ステージング環境

```sh:staging/cloudbuild.simplewebapp.yaml
steps:
- name: 'gcr.io/cloud-builders/gcloud'
  args: [
      'builds', 
      'submit',
      '--config=build/cloudbuild/_base/cloudbuild.simplewebapp.yaml',
      '--substitutions=_GCR_PROJECT=ds-abc123-staging',
      '.'
  ]
```

本番環境

```sh:prod/cloudbuild.simplewebapp.yaml
steps:
- name: 'gcr.io/cloud-builders/gcloud'
  args: [
      'builds', 
      'submit',
      '--config=build/cloudbuild/_base/cloudbuild.simplewebapp.yaml',
      '--substitutions=_GCR_PROJECT=ds-abc123-prod',
      '.'
  ]
```

## 4.　Cloudbuild実施
サンプルウェブアプリケーションのイメージビルドを行います。

```sh:開発環境
cd cloudbuild-multi-env-folder

# build image for dev env
gcloud builds submit --config build/cloudbuild/dev/cloudbuild.simplewebapp.yaml
```

```sh:ステージング環境
cd cloudbuild-multi-env-folder

# build image for staging env
gcloud builds submit --config build/cloudbuild/staging/cloudbuild.simplewebapp.yaml
```

```sh:本番環境
cd cloudbuild-multi-env-folder

# build image for prod env
gcloud builds submit --config build/cloudbuild/prod/cloudbuild.simplewebapp.yaml
```

<br>  
本記事で利用したソースコードはこちら
[https://github.com/dssolutioninc/dss_gke/tree/master/cloudbuild_multi_env](https://github.com/dssolutioninc/dss_gke/tree/master/cloudbuild_multi_env)

 
<br> 
最後まで読んで頂き、どうも有難う御座います!

DSS 橋本
