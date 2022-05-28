# GKEのIngressで複数サービスを１つのEndpoindにまとめ

## 　1．はじめに
マイクロサービスを導入すると、サービス数がどんどん多くなります。
サービスEndpoint単位で管理すると、システムが複雑となります。
Domain名の設定もめんどくさくなります。
Ingressを使用して、複数のサービスを１つのEndpointにまとめる構成を紹介したいと思います。


## 　2．アーキテクチャー
早速ですが、想像しやすくするため、全体アーキテクチャーを先に提示します。
![gke_ingress_for_collection_of_services_001.png](https://www.dssolution.jp/wp-content/uploads/2020/08/gke_ingress_for_collection_of_services_001.png)


これから構築に行きます。

## 　3．アプリケーションのフォルダ構成

```sh
application_folder
├── README.md
├── build
│   ├── cloudbuild.task_service.yaml
│   ├── cloudbuild.user_service.yaml
│   ├── task_service.Dockerfile
│   └── user_service.Dockerfile
├── deployment
│   └── application.deployment.yaml
├── go.mod
├── go.sum
├── openapiv2
│   ├── task_service
│   │   └── task_endpoint.yaml
│   └── user_service
│       └── user_endpoint.yaml
└── services
    ├── task
    │   ├── handler
    │   │   └── task_handler.go
    │   └── task.go
    └── user
        ├── handler
        │   └── user_handler.go
        └── user.go
```

## 　4．提供APIリスト
下記の2つURLを提供します。

```sh
https://api.example.com/task/v1/current   : 実地中のタスク内容取得
https://api.example.com/user/v1/roles     : ロールリストの取得
```

詳細な処理はソースコードをご参照。


## 　5．Container Registryにイメージビルド

Taskサービス

```sh
gcloud builds submit --config ingress/build/cloudbuild.task_service.yaml
```

Userサービス

```sh
gcloud builds submit --config ingress/build/cloudbuild.user_service.yaml
```


## 　6．Ingress定義で、リクエストに対して対象のサービスにルート
上記の構成図どおり、Ingressの後ろに２つのサービスがあります。
リクエストに対して、Userサービスか、Taskサービスかにルートします。
その設定はIngressの定義で制御します。

```yaml:application.deployment.yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: example-ingress
spec:
  rules:
  - http:
      paths:
      - path: /task/v1/*
        backend:
          serviceName: task-service
          servicePort: 80
      - path: /user/v1/*
        backend:
          serviceName: user-service
          servicePort: 80
---
...ソースコードにご参照
```

## 　7．GKEにデプロイ

```sh
kubectl apply -f deployment/application.deployment.yaml

# デプロイ後のIngress情報確認
kubectl get ingresses
> NAME              HOSTS       ADDRESS           PORTS     AGE
> example-ingress     *         203.0.113.123     80        59s
```


<br>  
本記事で利用したソースコードはこちら
[https://github.com/dssolutioninc/dss_gke/tree/master/ingress](https://github.com/dssolutioninc/dss_gke/tree/master/ingress)

<br> 
最後まで読んで頂き、どうも有難う御座います!

DSS 橋本