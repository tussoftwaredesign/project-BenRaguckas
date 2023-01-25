package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func init() {

}

func newDefaultMinioClient() *minio.Client {
	endpoint := "localhost:30036"
	accessKeyID := "RFeFTB2SJkK548Vb"
	secretAccessKey := "E3PR4OXJnluwSgJbWIN6Erfd4G9eibjb"
	useSSL := false
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		//	Exit 1 Unable to connect to minio
		os.Exit(1)
	}
	//	Avoid the error "# declared but never used"
	return minioClient
}

func testDefault(c *gin.Context) {
	c.String(http.StatusOK, fmt.Sprintf("default test api"))
}

func testSaveFile(c *gin.Context, identifier string) {
	file, err := c.FormFile(identifier)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to get form file: \n"+err.Error())
	}
	var save_path = "/usr/src/bin/localSave/"
	if os := runtime.GOOS; os == "windows" {
		save_path = "C:/Users/bean/project-y4/msgb/go-rest/localSave"
	}
	if err := c.SaveUploadedFile(file, save_path+file.Filename); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("'%s' failed to upload!", file.Filename))
	} else {
		c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
	}

}

func testMinioHealth(c *gin.Context) {
	minioClient := newDefaultMinioClient()
	cancelFn, err := minioClient.HealthCheck(5 * time.Second)

	if err == nil {
		c.String(http.StatusOK, "Success")
		defer cancelFn()
	} else {
		c.String(http.StatusInternalServerError, "Failed: "+err.Error())
	}
}

func testMinioAddFile(c *gin.Context, identifier string, bucket_name string) {
	minioClient := newDefaultMinioClient()
	fileHeaders, err := c.FormFile(identifier)
	if err != nil {
		c.String(http.StatusBadRequest, "No file found.")
	}
	file, err := fileHeaders.Open()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to read file provided: \n"+err.Error())
	}
	defer file.Close()
	f, err := minioClient.PutObject(c.Request.Context(), bucket_name, fileHeaders.Filename, file, fileHeaders.Size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to save file: \n"+err.Error())
	} else {
		c.String(http.StatusOK, fmt.Sprintf("File uploaded: \n\tbucket: %s\n\tkey: %s\n\tsize: %d", f.Bucket, f.Key, f.Size))
	}
}

func testMinioCreateBucket(c *gin.Context, bucket_name string) {
	minioClient := newDefaultMinioClient()
	err := minioClient.MakeBucket(context.Background(), bucket_name, minio.MakeBucketOptions{})
	if err != nil {
		fmt.Println(err)
		c.String(http.StatusInternalServerError, "Error in creating the bucket: \n"+err.Error())
		return
	}
	c.String(http.StatusOK, "Bucket created")

}

func testMinio(c *gin.Context) {
	ctx := context.Background()
	endpoint := "localhost:30036"
	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	useSSL := true

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	bucketName := "mymusic"
	location := "us-east-1"

	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}

	// Upload the zip file
	objectName := "golden-oldies.zip"
	filePath := "/tmp/golden-oldies.zip"
	contentType := "application/zip"

	// Upload the zip file with FPutObject
	info, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)
}
