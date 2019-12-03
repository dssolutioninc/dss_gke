# [Terraform](https://www.terraform.io/)ツールを使ってGCPリソース管理
Terraformのv0.12.16バージョンを使っています。（この記事記載時点の最新バージョンです）

本記事の目的
・Terraformを使ってGCPの環境を構築する時に、必要な設定のご紹介
・Terraformのよく使っているコマンドのご紹介


## 1. 必要な設定のご紹介
### 1.1 GCPプロジェクト設定
```sh
# GCPにログイン
gcloud auth login

# ログインブラウンザーが開かれて、自分のアカウントログインする。ログイン成功となったら、下記を続き

# 権限があるプロジェクトを全て表示
gcloud projects list

# ワーキングプロジェクト設定。プロジェクトリストから [PROJECT_ID]をコピーして下記のコマンドに入れる
gcloud config set project [PROJECT_ID]
```

### 1.2 利用するサービスのAPIを有効する
TerraformはAPIでGCPとやり取りするため、使うリソースに応じるサービスのAPIを有効する必要。

```sh
# 例：使うサービスのAPI有効
# 実施するアカウントは権限ある必要
gcloud services enable cloudresourcemanager.googleapis.com
gcloud services enable iam.googleapis.com
gcloud services enable compute.googleapis.com
gcloud services enable serviceusage.googleapis.com
gcloud services enable container.googleapis.com
gcloud services enable pubsub.googleapis.com
gcloud services enable storage-component.googleapis.com
```

### 1.3 Terraform専用のサービスアカウント作成・権限設定
TerraformはGCPにアクセスするため、アクセス用のCredentialが必要です。
通常の方法はTerraform専用のサービスアカウントを作成して、アカウントの権限付与をして、そのアカウントのCredentialを発行する。
Terraformの操作環境でCredentialを設定する。

```sh
# Terraform専用のサービスアカウント作成。アカウント名は「terraform-serviceaccount」とする
gcloud iam service-accounts create terraform-serviceaccount \
  --display-name "Account for Terraform"

# サービスアカウントに権限付与
# 使う範囲によりとなりますが、一旦「editor」ロールの権限を付与します。
# 下記コマンド中の[PROJECT_ID]に実際のプロジェクトIDを入れ替えてください。
gcloud projects add-iam-policy-binding [PROJECT_ID] \
  --member serviceAccount:terraform-serviceaccount@[PROJECT_ID].iam.gserviceaccount.com \
  --role roles/editor
```

### 1.4 サービスアカウントのCredentialファイルを作成して、Terraform稼働環境に設定
```sh
# そのアカウントのCredential発行
# Credentialのファイル名は「account.json」とする
gcloud iam service-accounts keys create path_to_save/account.json \
  --iam-account terraform-serviceaccount@[PROJECT_ID].iam.gserviceaccount.com

# 専用の環境変数にCredentialファイルを設定する
$ export GOOGLE_CLOUD_KEYFILE_JSON=path_to/account.json
```

## 2. Terraformのよく使っている使っているコマンドのご紹介
環境ごとのstateファイルは環境名ごとのフォルダで管理している。

### 2.1 Terraformスクリプト作成
```sh
# 初期化（初回のみ）
cd [TERRAFORM_FOLDER]

terraform init
```

```sh
# tfファイルを編集する
vi main.tf
```

```sh
# フォーマットする（tfファイルを編集時のみ）
terraform fmt
```

### 2.2 デプロイ前の差分確認
```sh
cd [TERRAFORM_FOLDER]

# tfファイルを適用する前に必ず差分を確認する
terraform plan
```

### 2.3 デプロイ実行
```sh
cd [TERRAFORM_FOLDER]

# planの結果が想定通りなら、tfファイルを適用する
terraform apply
```

### 2.4 その他、よく使うコマンド

```sh
cd [TERRAFORM_FOLDER]

# 何らかの理由で先にGCPへ物を作ってしまった場合、importでtfstateへ反映可能。
terraform import <tfファイルのリソース名> <GCPのリソース名>

# 例
terraform import google_storage_bucket.private-bucket syns-226007/asia-northeast1/private-bucket-abc123
```

```sh
# tfstateファイルを最新化したい
terraform refresh
```

## 3. Ref

* [GCP用に用意されたTerraformのドキュメント](https://www.terraform.io/docs/providers/google/)