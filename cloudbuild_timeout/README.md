# GCP CloudbuildでTimeoutエラーの対応

GCPのCloudbuildはタイムアウト時間のデフォルトが10分となります。
特に指定しない場合、10分以上かかるビルドはタイムアウトエラーとなります。

それを解決するため、timeoutだけを指定したら済みます。
ただし、いくつか注意はあります。

- timeoutは最大24時間（1日）で指定できます。
- Cloudbuildスクリプトがモジュール化される場合、モジュールごとに指定しないといけない。片方だけ指定する場合、残りのモジュールはタイムアウト10分のデフォルトのままとなります。

サンプルを投稿致します。

フォルダ構成

```sh
cloudbuild_timeout
├── _base
│   └── cloudbuild.matlab.R2019b.yaml
├── dev
│   └── cloudbuild.matlab.R2019b.yaml
└── dockerfile
    └── Dockerfile.matlab.R2019b
```

dev　→　_baseをコールする構成となっています。
この場合は両方のスクリプトで　timeoutを指定します。

```sh:dev/cloudbuild.matlab.R2019b.yaml
steps:
- name: 'gcr.io/cloud-builders/gcloud'
  args: [
      'builds', 
      'submit',
      '--config=_base/cloudbuild.matlab.R2019b.yaml',
      '--substitutions=_GCR_PROJECT=project-abc123,_GCR_REGION=asia.gcr.io',
      '.'
  ]
timeout: 3600s
```

```sh:_base/cloudbuild.matlab.R2019b.yaml
steps:

# docker build
- name: 'gcr.io/cloud-builders/docker'
  dir: .
  args: [
         'build',
         '-t', '${_GCR_REGION}/${_GCR_PROJECT}/${_GCR_IMAGE_NAME}:${_GCR_TAG}',
         '-f', 'dockerfile/Dockerfile.matlab.R2019b',
         '--cache-from', '${_GCR_REGION}/${_GCR_PROJECT}/${_GCR_IMAGE_NAME}:${_GCR_TAG}',
         '.'
        ]
        
- name: 'gcr.io/cloud-builders/docker'
  args: ["push", '${_GCR_REGION}/${_GCR_PROJECT}/${_GCR_IMAGE_NAME}']

timeout: 3600s

substitutions:
  # # Image name
  _GCR_IMAGE_NAME: matlab-r2019b
  # # Image tag
  _GCR_TAG: latest
  # # KMS Key location to decrypt private key

```


<br>
ご覧して頂き、どうも有難う御座います!
DevSamurai Ben