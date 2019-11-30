# [Terraform](https://www.terraform.io/)ツールを使ってGCPリソース管理
Terraformのv0.12.16バージョンを使っています。（この記事記載時点の最新バージョンです）

本記事の目的
・TerraformツールをGCP環境に接続するための設定
・Terraformのよく使っているコマンドのご紹介

記事の流れ
・事前準備（GCPプロジェクトの設定）
・Terraform専用のサービスアカウント作成・設定
・Terraformのよく使っている使っているコマンドのご紹介


## 事前準備
GCPプロジェクト設定
```sh
# GCPにログイン
gcloud auth login

# ログインブラウンザーが開かれて、自分のアカウントログインする。ログイン成功となったら、下記を続き

# 権限があるプロジェクトを全て表示
gcloud projects list

# ワーキングプロジェクト設定。プロジェクトリストから [PROJECT_ID]をコピーして下記のコマンドに入れる
gcloud config set project [PROJECT_ID]
```

## Terraform専用のサービスアカウント作成・設定
TerraformはGCPにアクセスするため、アクセス用のCredentialが必要です。
通常の方法はTerraform専用のサービスアカウントを作成して、アカウントの権限付与をして、そのアカウントのCredentialを発行する。
Terraformの操作環境でCredentialを設定する。

その手順
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

# そのアカウントのCredential発行
# Credentialのファイル名は「account.json」とする
gcloud iam service-accounts keys create path_to_save_folder/account.json \
  --iam-account terraform-serviceaccount@[PROJECT_ID].iam.gserviceaccount.com

# 専用の環境変数にCredentialファイルを設定する
$ export GOOGLE_CLOUD_KEYFILE_JSON=path_to/account.json
```

## Terraformのよく使っている使っているコマンドのご紹介
環境ごとのstateファイルは環境名ごとのフォルダで管理している。

```sh

```

```sh
$ cd terraform_folder

# 初期化（初回のみ）
$ terraform init
```

```sh
$ cd terraform_folder

# tfファイルを編集する
$ vi main.tf

```

```sh
$ cd terraform_folder

# フォーマットする（tfファイルを編集時のみ）
$ terraform fmt
```

```sh
$ cd terraform_folder

# tfファイルを適用する前に必ず差分を確認する
# 開発環境
$ terraform plan -var-file="dev.tfvars" -state=./dev/terraform.tfstate

# Staging環境
$ terraform plan -var-file="staging.tfvars" -state=./staging/terraform.tfstate

# 本番環境
$ terraform plan -var-file="prod.tfvars" -state=./prod/terraform.tfstate
```

```sh
$ cd terraform_folder

# planの結果が想定通りなら、tfファイルを適用する
# 開発環境
$ terraform apply -var-file="dev.tfvars" -state=./dev/terraform.tfstate

# Staging環境
$ terraform apply -var-file="staging.tfvars" -state=./staging/terraform.tfstate

# 本番環境
$ terraform apply -var-file="prod.tfvars" -state=./prod/terraform.tfstate
```

その他、よく使うコマンド

```sh
$ cd terraform_folder

# 何らかの理由で先にGCPへ物を作ってしまった場合、importでtfstateへ反映可能。
$ terraform import -var-file="dev.tfvars" -state=./dev/terraform.tfstate <tfファイルのリソース名> <GCPのリソース名>
```

```sh
$ cd terraform_folder

# tfstateファイルを最新化したい
$ terraform refresh -var-file="dev.tfvars" -state=./dev/terraform.tfstate
```

## Ref

* [GCP用に用意されたTerraformのドキュメント](https://www.terraform.io/docs/providers/google/)
