# GKE上にNFSを構築する方法


別の記事で[永続ディスク（Persistent Disk）](https://qiita.com/devs_hd/items/481a61f9e74f0f2758ec)設定方法を紹介しましたが、永続ディスクの制限は複数ノードからマウントして同時に読み書きできない。
本記事では、NFS（Network File System）を利用して、その制限を解決する方法を紹介致します。


実施手順

- Persistent Disk作成
- Persistent Diskフォーマット
- Persistent Diskを使って、NFSサーバ立ち上げ
- NFSを使って、GKE中にストレージを作成
- GKEのストレージマウントのPod作成

最初の２つのステップ（Persistent Disk作成、とPersistent Diskフォーマット）は別の記事で紹介したため、それらの[記事](https://qiita.com/devs_hd/items/481a61f9e74f0f2758ec)を参照してください。


※本手順はGCPとgcloudコマンドの利用経験があることが望ましいです。

ワークフォルダの構成

```sh
nfs_on_gke
├── README.md
├── cloudbuild.dummyjob.yaml
├── deployment
│   ├── dummy_job_01.deployment.yaml
│   └── dummy_job_02.deployment.yaml
├── dummyjob.Dockerfile
├── job
│   └── dummyjob.go
├── nfs-container.deployment.yaml
├── nfs-service.deployment.yaml
└── nfs-volume.yaml
```

## 1.  Persistent Diskを使って、NFSサーバ立ち上げ


 ```sh:nfs-container.deployment.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: nfs-server
spec:
  replicas: 1
  selector:
    matchLabels:
      role: nfs-server
  template:
    metadata:
      labels:
        role: nfs-server
    spec:
      containers:
      - name: nfs-server
        image: gcr.io/google_containers/volume-nfs:latest
        ports:
          - name: nfs
            containerPort: 2049
          - name: mountd
            containerPort: 20048
          - name: rpcbind
            containerPort: 111
        securityContext:
          privileged: true
        volumeMounts:
          - mountPath: /exports
            name: mypvc
      volumes:
        - name: mypvc
          gcePersistentDisk:
            pdName: gce-nfs-disk
            fsType: ext4
```

 ```sh:nfs-service.deployment.yaml
apiVersion: v1
kind: Service
metadata:
  name: nfs-server
spec:
  ports:
    - name: nfs
      port: 2049
    - name: mountd
      port: 20048
    - name: rpcbind
      port: 111
  selector:
    role: nfs-server
  type: LoadBalancer
```

GKEにデプロイ

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

# デプロイ NFSサーバ
kubectl apply -f nfs-container.deployment.yaml

# 外からアクセスするため、NFSサービスをデプロイ。クラスタ内でアクセスしかない場合、このステップをスキップ
kubectl apply -f nfs-service.deployment.yaml
```


デプロイ後の確認
<img width="910" alt="gcp_gke_kubernetes_nfs_dss_001.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/562088f2-5ffe-63d7-5734-996c4bc5aba7.png">
<img width="953" alt="gcp_gke_kubernetes_nfs_dss_002.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/a6e83928-f331-0b88-6a29-ccf092c23594.png">


## 2.  NFSを使って、GKE中にストレージを作成


```sh:nfs-volume.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs-data-volume
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteMany
  nfs:
    server: "xx.xx.xx.xx"
    path: "/exports"
---
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: nfs-data-volume
spec:
  accessModes:
    - ReadWriteMany
  storageClassName: ""
  resources:
    requests:
      storage: 10Gi
```

ボリューム定義実施

```sh
# PersistentVolumeとPersistentVolumeClaimを作成。複数接続で読み書きできるaccessModes（ReadWriteMany）
kubectl apply -f nfs-volume.yaml
```

定義実施後確認
<img width="947" alt="gcp_gke_kubernetes_nfs_dss_003.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/30bba43b-33fe-8776-f760-5ef4d0f271df.png">



## 3.  GKEのストレージマウントのPod作成

golangで簡単なジョブを作成する。
ジョブ処理はoutput-pathにサンプル１０ファイルを作成する。予定は複数ジョブを稼働して、共有用のNFSにファイルを作成する。

```go:dummyjob.go
package main

import (
    "flag"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "time"
)

var jobName = flag.String("job-name", "", "specify job name")
var outputPath = flag.String("output-path", "", "specify output path for job")

func main() {

    flag.Parse()

    outlog("Dummy Job start ...")

    for i := 0; i < 10; i++ {
        unixtime := time.Now().Unix()
        fileName := fmt.Sprint(*jobName, "_", unixtime, ".txt")

        // make a file
        file, err := os.Create(*outputPath + "/" + fileName)
        if err != nil {
            log.Fatal(err)
        }
        outlog("created a file: ", fileName)

        // out some
        file.WriteString("hello from " + *jobName)
        file.Close()

        // sleep to delay process
        time.Sleep(2 * time.Second)
    }

    // list all file in path
    outlog("List all files:")
    files, err := ioutil.ReadDir(*outputPath)
    if err != nil {
        log.Fatal(err)
    }

    for _, file := range files {
        outlog(file.Name())
    }

    outlog("Dummy Job finished.")

}

func outlog(args ...string) {
    log.Println(*jobName+":", args)
}
```


```sh:dummyjob.Dockerfile
FROM alpine:latest
WORKDIR /app
COPY ./dummyjob /app
```


```sh:cloudbuild.dummyjob.yaml
steps:
# go build
- name: golang:1.12
  dir: .
  args: ['go', 'build', '-o', 'dummyjob', 'job/dummyjob.go']
  env: ["CGO_ENABLED=0"]

# docker build
- name: 'gcr.io/cloud-builders/docker'
  dir: .
  args: [
         'build',
         '-t', '${_GCR_REGION}/${_GCR_PROJECT}/${_GCR_IMAGE_NAME}:${_GCR_TAG}',
         '-f', 'dummyjob.Dockerfile',
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
  _GCR_IMAGE_NAME: dummy-job
  # # Image tag
  _GCR_TAG: latest
```

イメージをビルドして、２つのジョブをデプロイする

```sh
# build dummy-job image on Container Registry
gcloud builds submit --config cloudbuild.dummyjob.yaml

# deploy job
kubectl apply -f deployment/dummy_job_01.deployment.yaml
kubectl apply -f deployment/dummy_job_02.deployment.yaml
```


ビルドイメージの確認
<img width="765" alt="gcp_gke_kubernetes_nfs_dss_004.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/d25b82e0-6eaa-3e3e-d983-398bff49c2a7.png">



デプロイジョブはGKEの中に確認
<img width="898" alt="gcp_gke_kubernetes_nfs_dss_005.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/a1da62ea-afb6-5da3-38e8-e8163af44b91.png">



２つのジョブはNFSを共有利用できることを稼働ログで確認できます。
<img width="766" alt="gcp_gke_kubernetes_nfs_dss_006.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/80f773e5-3f04-9336-6c9c-108ddfa44e84.png">



NFSを共有して読み書きできることを確認できました。

 

本記事で利用したソースコードはこちら

[https://github.com/dssolutioninc/dss_gke/tree/master/nfs](https://github.com/dssolutioninc/dss_gke/tree/master/nfs)

 

最後まで読んで頂き、どうも有難う御座います!

DSS 橋本