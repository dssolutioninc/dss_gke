package handler

import (
	"fmt"
	"os"
	"os/exec"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
)

type (
	SampleAppHandler struct {}

	StoregeFile struct {
		FileName string `json:"fileName" form:"fileName"`
		Content  string `json:"fileContent" form:"fileContent"`
	}
)

var (
	projectID string = os.Getenv("PROJECT_ID")
	serviceAccountKeyFile string = os.Getenv("ACCOUNT_JSON_FILE")
	storageBacketName string = os.Getenv("STORAGE_BUCKET_NAME")

	htmlHead string = `
<!DOCTYPE html><title>Demo gsutil on GKE</title>
<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/3.4.1/css/bootstrap.min.css" integrity="sha384-HSMxcRTRxnN+Bdg0JdbxYKrThecOKuH5zCYotlSAcp1+c8xmyTe9GYg1l9a69psu" crossorigin="anonymous">
<body>
	<div class="container">
		<h1 class="text-primary">Demo gsutil on GKE</h1>
		<hr>

		<form method='POST' action='/createfile'>
			<div class="form-group">
				<label for="fileName">File Name</label>
				<input type="text" class="form-control" name="fileName" placeholder="New File Name">
			</div>
			<div class="form-group">
				<label for="fileContent">File Content</label>
				<textarea name="fileContent" class="form-control" rows="3"></textarea>
			</div>
			<button type="submit" class="btn btn-default">Submit</button>
		</form>
	<hr>
	<h2 class="text-primary">List files in your Bucket</h2>
	`

	htmlTail string = `
	</div>
	<script src="https://stackpath.bootstrapcdn.com/bootstrap/3.4.1/js/bootstrap.min.js" integrity="sha384-aJ21OjlMXNL5UyIl/XNwTMqvzeRMZH2w8c5cRVpzpU8Y5bApTppSuUkhZXN0VxHd" crossorigin="anonymous"></script>
</body>
</html>
`
)

func (sh SampleAppHandler) Index(c echo.Context) error {
	err := initGsutil(serviceAccountKeyFile, projectID)
	if err != nil {
		return err
	}

	cmd := exec.Command("gsutil", "ls", "-r", storageBacketName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	resHtml := fmt.Sprintf("<pre><code>%s</code></pre>", out)

	return c.HTML(http.StatusOK, htmlHead + resHtml + htmlTail)
}

func (sh SampleAppHandler) CreateFile(c echo.Context) error {
	// get data in form request 
	storegeFile := new(StoregeFile)
	if err := c.Bind(storegeFile); err != nil {
		return err
	}

	// create file
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

	// init to use gsutil
	err = initGsutil(serviceAccountKeyFile, projectID)
	if err != nil {
		return err
	}

	// copy file to bucket
	cmd := exec.Command("gsutil", "cp", tmpfile.Name(), storageBacketName+"/"+storegeFile.FileName)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
	// return sh.Index(c)
}

func initGsutil(serviceAccountKeyFile string, projectID string) error {
	cmd := exec.Command("gcloud", "auth", "activate-service-account", "--key-file="+serviceAccountKeyFile)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	cmd = exec.Command("gcloud", "config", "set", "project", projectID)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}