package main

import (
	"context"
	"fmt"
	"mime/multipart"
	"strconv"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func getMinioClient() (*minio.Client, error) {
	fmt.Printf("MINIO: %s:%d\n", minio_host, minio_port)
	return minio.New(minio_host+":"+strconv.Itoa(minio_port), &minio.Options{
		Creds:  credentials.NewStaticV4(minio_cred_id, minio_cred_key, ""),
		Secure: false,
	})
}

func getMinioBucketItems(bucket_uuid string) (<-chan minio.ObjectInfo, error) {
	minioClient, err := getMinioClient()
	if err != nil {
		return nil, err
	}
	objects := minioClient.ListObjects(context.Background(), bucket_uuid, minio.ListObjectsOptions{})
	return objects, nil
}

func getMinioBucketItem(bucker_uuid string, item_name string) (*minio.Object, error) {
	minioClient, err := getMinioClient()
	if err != nil {
		return nil, err
	}
	return minioClient.GetObject(context.Background(), bucker_uuid, item_name, minio.GetObjectOptions{})
}

func pushMinioFile(file multipart.File, fileHeaders multipart.FileHeader) (*minio.UploadInfo, error) {
	defer file.Close()
	bucket_uuid := uuid.New().String()
	minioClient, err := getMinioClient()
	if err != nil {
		return nil, err
	}
	err = minioClient.MakeBucket(context.Background(), bucket_uuid, minio.MakeBucketOptions{})
	if err != nil {
		return nil, err
	}
	upload_info, err := minioClient.PutObject(
		context.Background(), //	Context (would it be better to use gin.Context ?)
		bucket_uuid,          //	Bucket name (use uuid)
		fileHeaders.Filename, //	File name (need potential system for this)
		file,                 //	Actual file
		fileHeaders.Size,     //	File size (for minio to optimize the upload, will need to inspect how it handles large files)
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	return &upload_info, err
}

func putMinioFile(file multipart.File, fileHeaders multipart.FileHeader, bucket_uuid string) (*minio.UploadInfo, error) {
	minioClient, err := getMinioClient()
	if err != nil {
		return nil, err
	}
	defer file.Close()
	upload_info, err := minioClient.PutObject(
		context.Background(), //	Context (would it be better to use gin.Context ?)
		bucket_uuid,          //	Bucket name (use uuid)
		fileHeaders.Filename, //	File name (need potential system for this)
		file,                 //	Actual file
		fileHeaders.Size,     //	File size (for minio to optimize the upload, will need to inspect how it handles large files)
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	return &upload_info, err
}

func putMinioFileNamed(file multipart.File, fileHeaders multipart.FileHeader, bucket_uuid string, item_name string) (*minio.UploadInfo, error) {
	fileHeaders.Filename = item_name
	fmt.Println(fileHeaders.Filename)
	return putMinioFile(file, fileHeaders, bucket_uuid)
}

func deleteMinioFile(bucker_uuid string, item_name string) error {
	minioClient, err := getMinioClient()
	if err != nil {
		return err
	}
	err = minioClient.RemoveObject(
		context.Background(),
		bucker_uuid,
		item_name,
		minio.RemoveObjectOptions{},
	)
	return err
}
