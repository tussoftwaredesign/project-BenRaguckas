package main

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// Returns minio client with default or cmd flag variables

// Minio Health Check call
func healthMinio(c *gin.Context) {
	minioClient, err := getMinioClient()
	//	Exclusive error response
	if err != nil {
		c.String(http.StatusInternalServerError, "Server failed to establish minio connection:\n"+err.Error())
	} else {
		cancelFn, err := minioClient.HealthCheck(3 * time.Second)
		//	Exclusive error response
		if err != nil {
			c.String(http.StatusInternalServerError, "Minio healthcheck failed:\n"+err.Error())
		} else {
			c.String(http.StatusOK, "Minio health check successful.")
			defer cancelFn()
		}
	}
}

// Gets file from form of given GinContext, returns errors if it failed to receive or open
func (c GinContext) getFormFile() (multipart.File, multipart.FileHeader) {
	fileHeaders, err := c.FormFile(form_file_identifier)
	//	Exclusive error response
	if err != nil {
		c.String(http.StatusBadRequest, "No file uploaded.")
		return nil, *fileHeaders
	}
	file, err := fileHeaders.Open()
	//	Exclusive error response
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to read file:\n"+err.Error())
		return nil, *fileHeaders
	}
	return file, *fileHeaders
}

func addNewMinioFile(c *gin.Context) {
	minioClient, err := getMinioClient()
	//	Exclusive error response
	if err != nil {
		c.String(http.StatusInternalServerError, "Server failed to establish minio connection:\n"+err.Error())
		return
	}
	//	Use extended method and return if unable to retrieve file
	r := &GinContext{c}
	file, fileHeaders := r.getFormFile()
	if file == nil {
		return
	}
	// Create new bucket with uuid as string
	bucket_uuid := uuid.New().String()
	err = minioClient.MakeBucket(context.Background(), bucket_uuid, minio.MakeBucketOptions{})
	//	Exclusive error response
	if err != nil {
		c.String(http.StatusInternalServerError, "Error in creating a bucket:\n"+err.Error())
		return
	}
	defer file.Close()
	upload_info, err := minioClient.PutObject(
		c.Request.Context(),  //	Context (would it be better to use gin.Context ?)
		bucket_uuid,          //	Bucket name (use uuid)
		fileHeaders.Filename, //	File name (need potential system for this)
		file,                 //	Actual file
		fileHeaders.Size,     //	File size (for minio to optimize the upload, will need to inspect how it handles large files)
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	//	Exclusive error response
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to save file: \n"+err.Error())
		return
	}
	//	ADD ranbbitMQ routing HERE
	//	Success response
	c.String(http.StatusOK, fmt.Sprintf("File uploaded: \n\tbucket: %s\n\tkey: %s\n\tsize: %d", upload_info.Bucket, upload_info.Key, upload_info.Size))

}

func putMinioFile2(c *gin.Context, bucket_uuid string) {
	minioClient, err := getMinioClient()
	//	Exclusive error response
	if err != nil {
		c.String(http.StatusInternalServerError, "Server failed to establish minio connection:\n"+err.Error())
		return
	}
	//	Use extended method and return if unable to retrieve file
	r := &GinContext{c}
	file, fileHeaders := r.getFormFile()
	if file == nil {
		return
	}
	defer file.Close()
	upload_info, err := minioClient.PutObject(
		c.Request.Context(),  //	Context (would it be better to use gin.Context ?)
		bucket_uuid,          //	Bucket name (use uuid)
		fileHeaders.Filename, //	File name (need potential system for this)
		file,                 //	Actual file
		fileHeaders.Size,     //	File size (for minio to optimize the upload, will need to inspect how it handles large files)
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	//	Exclusive error response
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to save file: \n"+err.Error())
		return
	}
	//	Success response
	c.String(http.StatusOK, fmt.Sprintf("File uploaded: \n\tbucket: %s\n\tkey: %s\n\tsize: %d", upload_info.Bucket, upload_info.Key, upload_info.Size))
}

func getMinioFile(c *gin.Context, bucket_uuid string, object_name string) {
	minioClient, err := getMinioClient()
	//	Exclusive error response
	if err != nil {
		c.String(http.StatusInternalServerError, "Server failed to establish minio connection:\n"+err.Error())
		return
	}
	file, err := minioClient.GetObject(context.Background(), bucket_uuid, object_name, minio.GetObjectOptions{})
	//	Exclusive error response (This does not fail.. ever...)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to get File: \n"+err.Error())
		return
	}
	stat, err := file.Stat()
	//	Exclusive error response
	if err != nil {
		c.String(http.StatusInternalServerError, "File stat error: \n"+err.Error()+object_name)
		return
	}
	defer file.Close()
	c.DataFromReader(http.StatusOK, stat.Size, "application/octet-stream", file, map[string]string{"Content-Disposition": `'attachment; filename="image.png"`})
}
