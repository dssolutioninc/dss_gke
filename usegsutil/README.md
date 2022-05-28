# GKE中のGolangアプリケーション、gsutilを使ってCloud Storageでデータ連携を行う

GKEの中に稼働されるアプリケーションからどうやってGCPサービスにアクセスしたり、データ連携したりするか？という疑問がある方々に回答する記事をまとめました。
今回はCloud Storageサービスを使って、データ連携サンプルとして作成しました。

APIを利用して、Cloud Storage上のデータファイルと遣り取り方法があります。この方法の利点は処理スピードが良いことです。
但し、エラーハンドリングや処理ロジックなどを全て実装しないといけない。

それより簡単な方法はgsutilツールを使ってCloud Storageと遣り取りする方法があります。
gsutil は、コマンドラインから Cloud Storage にアクセスできる Python アプリケーションです。
Cloud Storageと遣り取り専用の正式なツールです。
gsutil を使用すると、次のような、バケットやオブジェクトの幅広い管理作業を行うことができます。

- バケットの作成と削除
- オブジェクトのアップロード、ダウンロード、削除
- バケットとオブジェクトの一覧表示
- オブジェクトの移動、コピー、名前変更
- オブジェクトやバケットの ACL の編集
- エラーハンドリングや平行

この記事はGKEにデプロイするアプリケーションはgsutilを利用するため、実装方法をサンプルとして紹介したいと思っています。

## 1.　はじめに
ローカル環境でこの流れを実施したら、gsutilを使えるようになります。

グーグルクラウドに接続するため、ログインを行います。
gcloud auth login

作業対象のプロジェクトを設定
gcloud config set project [PROJECT_ID]

この状態で、gsutilを使う可能となります。例えば、ストレージのバケツを全て表示するコマンド
gsutil ls -r gs://[BUCKET_NAME]

GKE中のアプリケーションの場合どうなるか？
回答は基本的にこの流れと同じです。但し、実装方法は違い点があります。
次、各項目で実装方法を紹介いたします。

## 2.　グーグルクラウドに接続ための設定
サービスカウントを使って接続します。

まず、サービスカウント作成と接続用のaccount.jsonファイル発行を行います。
詳細の実施方法はこの記事の「[サービスアカウント作成＆アクセス用のaccount.jsonファイル発行](GKE中のGolangアプリケーションからCloud Pub/Subを使ってデータ連携を行う
https://qiita.com/devs_hd/items/83853cb83128862db9d6#1%E3%82%B5%E3%83%BC%E3%83%93%E3%82%B9%E3%82%A2%E3%82%AB%E3%82%A6%E3%83%B3%E3%83%88%E4%BD%9C%E6%88%90%E3%82%A2%E3%82%AF%E3%82%BB%E3%82%B9%E7%94%A8%E3%81%AEaccountjson%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB%E7%99%BA%E8%A1%8C)」
で書いていますので、ご参照をお願いします。

それから、account.jsonファイルを使ってグーグルクラウドに接続設定はこのコマンドで実施

```sh
gcloud auth activate-service-account --key-file=[ACOUNT_JSON_FILE_PATH]
```

## 3.　サンプルのアプリケーション実装
下記2つ機能のアプリケーションを実装に行きます。

- ファイル名とテキスト内容を指定してファイルをCloud Storageのバケツ中に作成する。
- バケツ中のファイルを全てリストする。

golangとechoフレームワークを使います。

### 　プログラムからグーグルクラウドに接続設定
execパケージを使って、gcloudツールで接続を行います。

```go
md := exec.Command("gcloud", "auth", "activate-service-account", "--key-file="+serviceAccountKeyFile)
_, err := cmd.CombinedOutput()
if err != nil {
	return err
}

cmd = exec.Command("gcloud", "config", "set", "project", projectID)
_, err = cmd.CombinedOutput()
if err != nil {
	return err
}
```

### 　ストレージのバケツにファイル作成
ioutil.TempFileで一時的にファイルを作成します。

```go
tmpfile, err := ioutil.TempFile("", storegeFile.FileName)
if err != nil {
	return err
}
defer os.Remove(tmpfile.Name()) // clean up

if _, err := tmpfile.Write([]byte(storegeFile.Content)); err != nil {
	return err
}
if err := tmpfile.Close(); err != nil {
	return err
}
```

それから、gsutilでバケツにコピーします。

```go
cmd := exec.Command("gsutil", "cp", tmpfile.Name(), storageBacketName+"/"+storegeFile.FileName)
_, err = cmd.CombinedOutput()
if err != nil {
	return err
}
```

### 　ストレージのバケツのファイルを全て表示
gsutilツールでファイルリストを表示します。

```go
cmd := exec.Command("gsutil", "ls", "-r", storageBacketName)
out, err := cmd.CombinedOutput()
if err != nil {
	return err
}
```


## 4.　ビルド＆デプロイ
### 　Dockerイメージ
gcloudとgsutilを使うため、Dockerイメージはgoogle/cloud-sdkを使います。
alpineを使ってイメージのサイズを最低化とします。

```sh
FROM google/cloud-sdk:alpine
WORKDIR /app
COPY ./sample_app /app

ENTRYPOINT ["./sample_app"]
```

### 　ビルド＆デプロイ
手順まとめ

```sh
# グーグルクラウドのContainer Registryにイメージビルド
gcloud builds submit --config cloudbuild.sampleapp.yaml

# GKEのクラスタにアクセスするため、credentialsを設定
gcloud container clusters get-credentials --zone asia-northeast1-b ds-gke-small-cluster

# サービスアカウントJSONファイルを利用するため、secretボリュームを作成
kubectl create secret generic service-account-credential \
     --from-file=./secret/account.json

# アプリケーションをGKEにデプロイする
kubectl apply -f deployment.sampleapp.yaml
```

デプロイ後、GKE中のアプリケーションを確認
<img width="1092" alt="gke_kubernetes_gsutil_ds_001.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/e4c952f2-8e9c-fbb1-6b6b-1c333c371c26.png">


EndpointsのIPをメモして、アプリケーションにアクセスと稼働検証します。
<img width="1177" alt="gke_kubernetes_gsutil_ds_002.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/188468a9-c105-d762-4911-ec7e4c03760b.png">


<br>  
本記事で利用したソースコードはこちら
[https://github.com/dssolutioninc/dss_gke/tree/master/usegsutil](https://github.com/dssolutioninc/dss_gke/tree/master/usegsutil)

 
<br> 
最後まで読んで頂き、どうも有難う御座います!
DSS 橋本

<br>
*関連記事*
[GKE中のGolangアプリケーションからCloud Pub/Subを使ってデータ連携を行う](https://qiita.com/devs_hd/items/83853cb83128862db9d6)