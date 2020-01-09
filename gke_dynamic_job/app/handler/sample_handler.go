package handler

import (
	"os"
	"fmt"
	"net/http"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"

	"github.com/labstack/echo/v4"
)

var (
	projectID string 				= os.Getenv("PROJECT_ID")
	containerRegistryRepo string 	= os.Getenv("CONTAINER_REGISTRY_REPO")
	statusUpdateEndpoint string 	= os.Getenv("STATUS_UPDATE_ENDPOINT")
)

type (
	SampleHandler struct {
	}

	JobInfo struct {
		JobId   string `json:"job-id" form:"job-id" query:"job-id"`
		JobName string `json:"job-name"`
		Status  string `json:"status" form:"status" query:"status"`
	}
)

var (
	jobs = []*JobInfo{}
)

func (jh SampleHandler) Index(c echo.Context) error {
	resHtml := `
		<!DOCTYPE html><title>Demo Run Kubernetes Job from Program</title>
		<head>
			<meta http-equiv="refresh" content="7" >
			<style>
				span {
					display: inline-block;
					width: 250px;
				}
			</style>
		</head>
		<h1>Demo Run Kubernetes Job from Program</h1>

		<form method='POST' action='/runajob'>
			<label> Job Description</label>
			<input required name='job-id' placeholder='Job ID'>
			<input required name='job-name' placeholder='Job Name'>
			<input type='submit' value='Run'>
		</form>

		<p>Job List</p>
			<ol>
	`

	for _, item := range jobs {
		resHtml += "<li><span>" + item.JobId +
			"</span><span>" + item.JobName +
			"</span><span>" + item.Status +
			"</span></li>"
	}
	resHtml += "</ol>"

	return c.HTML(http.StatusOK, resHtml)
}

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

func int32Ptr(i int32) *int32 { return &i }
