co# [Terraform](https://www.terraform.io/)を利用したインフラ定義


## 事前準備

サービスアカウントのaccount.jsonを準備して、配置すること。

```sh
# create service account
gcloud iam service-accounts create terraform-serviceaccount \
  --display-name "Account for Terraform"

# NEED ADMIN to add permission
# example for staging env
gcloud config set project syns-sol-grdsys-stage

gcloud projects add-iam-policy-binding syns-sol-grdsys-stage \
  --member serviceAccount:terraform-serviceaccount@syns-sol-grdsys-stage.iam.gserviceaccount.com \
  --role roles/editor

# create credentials file account.json
gcloud iam service-accounts keys create path_to_save_folder/account.json \
  --iam-account terraform-serviceaccount@syns-sol-grdsys-stage.iam.gserviceaccount.com
```


## Terraform　操作
環境ごとのstateファイルは環境名ごとのフォルダで管理している。

```sh
# set credentials to environment variable
$ export GOOGLE_CLOUD_KEYFILE_JSON={{path_to_account.json}}
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
