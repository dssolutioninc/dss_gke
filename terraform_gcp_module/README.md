# [Terraform](https://www.terraform.io/)スクリプトをモジュール化して、GCPの複数環境に適用
※Terraformのv0.12.16バージョンを使っています。（この記事記載時点の最新バージョンです）

本記事の目的
・Terraformスクリプトをモジュール化して、GCPの開発環境、テスト環境、本番環境に適用する方法のご紹介

Terraformは初めての方はこの記事（[Terraformツールを使ってGCPリソース管理](https://qiita.com/devs_hd/items/6a715fedf5462af420f2)）もご覧ください。


## 1. 　Terraformスクリプト作成
この記事では、GKEクラスタ、ストレージBucket、Pubsub Topic＆Subscriptionを例としてデプロイスクリプトを作成します。

Terraformスクリプトフォルダの構成

```sh
terraform_script_folder
├── _modules
│   ├── cluster
│   │   ├── main.tf
│   │   ├── outputs.tf
│   │   └── variables.tf
│   ├── pubsub
│   │   ├── main.tf
│   │   ├── outputs.tf
│   │   └── variables.tf
│   └── storage
│       ├── main.tf
│       ├── outputs.tf
│       └── variables.tf
├── dev
│   ├── account.json
│   └── terraform.tfstate
├── dev.tfvars
├── main.tf
├── prod
│   ├── account.json
│   └── terraform.tfstate
├── prod.tfvars
├── staging
│   ├── account.json
│   └── terraform.tfstate
├── staging.tfvars
└── variables.tf
```

フォルダ構成の説明

- _modulesフォルダ：リソース種類のごとに共有定義スクリプトを格納する
- dev、staging、prodのフォルダ：開発、ステージング、本番の環境用にアクセス用のアカウントファイルとStateファイルを格納
- dev.tfvars、staging.tfvars、prod.tfvarsのファイル：各種環境によるパラメータ設定ファイル


## 2.  環境別にデプロイ実施
#### 2.1  開発環境

```sh
# 専用の環境変数にCredentialファイルを設定する
$ export GOOGLE_CLOUD_KEYFILE_JSON=path_to/dev/account.json

# tfファイルを適用する前に必ず差分を確認する
cd [TERRAFORM_FOLDER]
terraform plan -var-file="dev.tfvars" -state=./dev/terraform.tfstate

# planの結果が想定通りなら、tfファイルを適用する
terraform apply -var-file="dev.tfvars" -state=./dev/terraform.tfstate
```

#### 2.2  ステージング環境

```sh
# 専用の環境変数にCredentialファイルを設定する
$ export GOOGLE_CLOUD_KEYFILE_JSON=path_to/staging/account.json

# tfファイルを適用する前に必ず差分を確認する
cd [TERRAFORM_FOLDER]
terraform plan -var-file="staging.tfvars" -state=./staging/terraform.tfstate

# planの結果が想定通りなら、tfファイルを適用する
terraform apply -var-file="staging.tfvars" -state=./staging/terraform.tfstate
```

#### 2.3  本番環境

```sh
# 専用の環境変数にCredentialファイルを設定する
$ export GOOGLE_CLOUD_KEYFILE_JSON=path_to/prod/account.json

# tfファイルを適用する前に必ず差分を確認する
cd [TERRAFORM_FOLDER]
terraform plan -var-file="staging.tfvars" -state=./prod/terraform.tfstate

# planの結果が想定通りなら、tfファイルを適用する
terraform apply -var-file="staging.tfvars" -state=./prod/terraform.tfstate
```

本記事の利用ソースコードはこちら
[https://github.com/itdevsamurai/gke/tree/master/terraform_gcp_module](https://github.com/itdevsamurai/gke/tree/master/terraform_gcp_module)


最後まで読んで頂き、どうも有難う御座います!
DevSamurai 橋本


関連記事：[Terraformツールを使ってGCPリソース管理](https://qiita.com/devs_hd/items/6a715fedf5462af420f2)