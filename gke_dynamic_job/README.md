# GKE中のプログラムからKubernetesジョブを生成して非同期処理を行う


## 1.　はじめに
定期的に稼働されるジョブではなく、リクエストされたらジョブを生成して処理を行いたい時はどの形で実装するか？
ジョブの稼働ステータスは何の方法で連携するか？疑問がある方々がいると思います。
この記事では実装方法のサンプルを紹介したいと思います。

早速ですが実装に行くサンプルアプリケーションの機能は
ウェブ画面からKubernetesジョブ実施を指示します。
実施ジョブ一覧とジョブ処理ステータスを表示させます。


## 2.　アプリケーションフォルダの構成
全体のイメージを想像しやすくなるため、アプリケーションのフォルダ構成を先に見せます。

```sh
gke_dynamic_job
├── README.md
├── app
│   ├── handler
│   │   └── sample_handler.go
│   └── sample_app.go
├── build
│   ├── cloudbuild.dummyjob.yaml
│   ├── cloudbuild.sampleapp.yaml
│   ├── dummyjob.Dockerfile
│   └── sampleapp.Dockerfile
├── deployment.sampleapp.yaml
├── doc
├── go.mod
├── go.sum
└── job
    └── dummyjob.go
```

## 3.　Kubernetesジョブを生成方法
KubernetesのGoクライアントを利用して、プログラムの中からジョブを作成と稼働させます。
ジョブの設定は色々ありますが、プログラムの中に全て指定します。
固定ではない設定は環境変数を経由にプログラムに渡します。


利用するライブラリーはKubernetesとGoクライアントのライブラリー

```sh
import (
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
)
```

ジョブ生成の関数

```sh
// Run a job
func (jh SampleHandler) RunAJob(c echo.Context) error {
	jobInfo := new(JobInfo)
	if err := c.Bind(jobInfo); err != nil {
		c.JSON(http.StatusBadRequest, "input parameters error")
	}

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	jobClient := clientset.BatchV1().Jobs("default")

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "dummy-job-",
		},
		Spec: batchv1.JobSpec{
			// the retries number when job failed. No retry for this dummy job
			BackoffLimit: int32Ptr(0),

			// auto delete job after finished
			TTLSecondsAfterFinished: int32Ptr(300),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "dummy-job-",
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:    "dummy-job-container",
							Image:   fmt.Sprintf("%s/%s/dummyjob", containerRegistryRepo, projectID),
							Command: []string{"/app/dummyjob"},
							Args:    []string{"--api-end-point", statusUpdateEndpoint, "--job-id", jobInfo.JobId},
						},
					},
					RestartPolicy: apiv1.RestartPolicyNever,
				},
			},
		},
	}

	fmt.Println("Creating job... ")

	result1, err1 := jobClient.Create(job)
	if err != nil {
		fmt.Println(err1)
		panic(err1)
	}

	jobName := result1.GetName()
	fmt.Printf("Created job %s\n", jobName)

	jobs = append(jobs, &JobInfo{
		JobId:      jobInfo.JobId,
		JobName: jobName,
		Status:  "-",
	})

	return c.Redirect(http.StatusMovedPermanently, "/index")
}
```

## 4.　Kubernetesジョブ稼働ステータスや結果の連携
別の記事で[クラウドストレージ](https://qiita.com/devs_hd/items/43fbd4603bb4ac143432)や[Pub/Sub](https://qiita.com/devs_hd/items/83853cb83128862db9d6)を利用して処理データ連携をご紹介いたしましたが、
本記事はウェブAPIを使って処理結果を連携するとします。

ウェブアプリケーションはジョブ処理結果を受け取りする専用のAPIを作成します。
ジョブの処理の中に、処理結果更新ウェブAPIをコールして、処理結果を更新します。


処理結果を受け取りする専用のAPIの関数

```sh
// Update job result
func (jh SampleHandler) UpdateJobStatus(c echo.Context) error {

	dataStatus := new(JobInfo)
	if err := c.Bind(dataStatus); err != nil {
		c.JSON(http.StatusBadRequest, "input parameters error")
	}

	fmt.Printf("JobId : %s\n", dataStatus.JobId)
	fmt.Printf("Status : %s\n", dataStatus.Status)

	for i, item := range jobs {
		if item.JobId == dataStatus.JobId {
			jobs[i].Status = dataStatus.Status

			break
		}
	}

	return c.JSON(http.StatusOK, "ok")
}
```

ジョブの中に、処理結果更新ウェブAPIをコールする

```sh
// update process status
http.PostForm(apiEndPoint, url.Values{"job-id": {jobID}, "status": {"ERROR"}})
```

## 5.　ビルド＆デプロイの手順

ジョブの Container Registry イメージビルド

```sh
gcloud builds submit --config build/cloudbuild.dummyjob.yaml
```

ウェブアプリケーションの Container Registry イメージビルド

```sh
gcloud builds submit --config build/cloudbuild.sampleapp.yaml
```

ウェブアプリケーションをGKEクラスタにデプロイ

```sh
kubectl apply -f deployment.sampleapp.yaml
```

デプロイ後の確認
<img width="954" alt="gcp_kubernetes_dynamic_job_001.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/09398470-793f-e57c-f0d4-ee3f6b85af3d.png">

<img width="1024" alt="gcp_kubernetes_dynamic_job_002.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/f59b9198-7216-7c06-9aac-252b791dfb3b.png">

Endpointsの IPをメモして、ウェブブラウザでウェブアプリケーションを開きます。
例
http://34.84.171.50/index

## 6.　cluster-admin権限付与
プログラムからクラスタの中にジョブを生成するため、稼働アカウントにcluster-admin権限を付与する必要です。
稼働アカウントは特に指定しない場合、defaultのアカウントとなります。
また namespace も指定しない場合、defaultのnamespace　の中に稼働となります。

cluster-admin権限付与の手順。
権限付与のため、owner権限アカウントで実施する必要です。

```sh
export WORK_PROJECT_ID=project-abc123
export WORK_ZONE=asia-northeast1-a

gcloud config set project ${WORK_PROJECT_ID}
gcloud container clusters get-credentials --zone ${WORK_ZONE} [YOUR_CLUSTER_NAME]
# assign cluster-admin role inside "default" namespace for default service account
# need admin account to run
kubectl create clusterrolebinding cluster-admin-permission-binding \
  --clusterrole=cluster-admin \
  --user=system:serviceaccount:default:default \
  --namespace=default
```


## 7.　稼働検証

ウェブアプリケーションのURLを開きます。
http://34.84.171.50/index

Job ID と Job Nameを指定して実施します。

<img width="763" alt="gcp_kubernetes_dynamic_job_005.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/f02210f6-7a69-e00e-739c-cc42fa18a5ed.png">

GKE中の生成されたジョブを確認に行きます。
<img width="1097" alt="gcp_kubernetes_dynamic_job_006.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/56d17e80-7b55-aa7f-7215-6cb8523aa25e.png">

バッチジョブのロジックはRandomで正常終了または異常終了となります。<br>
１つジョブの稼働ログ。
<img width="1145" alt="gcp_kubernetes_dynamic_job_007.png" src="https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/535698/190efce8-0310-f03d-db56-536bc36e85d8.png">

サンプルアプリケーションの実装はここで完了となります。


本記事で利用したソースコードはこちら
[https://github.com/dssolutioninc/dss_gke/tree/master/gke_dynamic_job](https://github.com/dssolutioninc/dss_gke/tree/master/gke_dynamic_job)


最後まで読んで頂き、どうも有難う御座います!
DSS 橋本