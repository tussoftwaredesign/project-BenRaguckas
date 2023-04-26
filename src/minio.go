package main

import (
	"context"
	"errors"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

var minioClient *minio.Client

// func getMinioBucketItems(bucket_uuid string) (<-chan minio.ObjectInfo, error) {
// 	objects := minioClient.ListObjects(context.Background(), bucket_uuid, minio.ListObjectsOptions{})
// 	return objects, nil
// }

// func getMinioBucketItem(bucker_uuid string, item_name string) (*minio.Object, error) {
// 	return minioClient.GetObject(context.Background(), bucker_uuid, item_name, minio.GetObjectOptions{})
// }

// func pushMinioFileOld(file multipart.File, fileHeaders multipart.FileHeader) (*minio.UploadInfo, error) {
// 	defer file.Close()
// 	bucket_uuid := uuid.New().String()
// 	err := minioClient.MakeBucket(context.Background(), bucket_uuid, minio.MakeBucketOptions{})
// 	if err != nil {
// 		return nil, err
// 	}
// 	upload_info, err := minioClient.PutObject(
// 		context.Background(), //	Context (would it be better to use gin.Context ?)
// 		bucket_uuid,          //	Bucket name (use uuid)
// 		fileHeaders.Filename, //	File name (need potential system for this)
// 		file,                 //	Actual file
// 		fileHeaders.Size,     //	File size (for minio to optimize the upload, will need to inspect how it handles large files)
// 		minio.PutObjectOptions{ContentType: "application/octet-stream"},
// 	)
// 	return &upload_info, err
// }

// func putMinioFile(file multipart.File, fileHeaders multipart.FileHeader, bucket_uuid string) (*minio.UploadInfo, error) {
// 	defer file.Close()
// 	upload_info, err := minioClient.PutObject(
// 		context.Background(), //	Context (would it be better to use gin.Context ?)
// 		bucket_uuid,          //	Bucket name (use uuid)
// 		fileHeaders.Filename, //	File name (need potential system for this)
// 		file,                 //	Actual file
// 		fileHeaders.Size,     //	File size (for minio to optimize the upload, will need to inspect how it handles large files)
// 		minio.PutObjectOptions{ContentType: "application/octet-stream"},
// 	)
// 	return &upload_info, err
// }

// func putMinioFileNamedOld(file multipart.File, fileHeaders multipart.FileHeader, bucket_uuid string, item_name string) (*minio.UploadInfo, error) {
// 	fileHeaders.Filename = item_name
// 	fmt.Println(fileHeaders.Filename)
// 	return putMinioFile(file, fileHeaders, bucket_uuid)
// }

// func deleteMinioFile(bucker_uuid string, item_name string) error {
// 	err := minioClient.RemoveObject(
// 		context.Background(),
// 		bucker_uuid,
// 		item_name,
// 		minio.RemoveObjectOptions{},
// 	)
// 	return err
// }

// ---
// ---
// Definitive function list
func minioPushFile(file multipart.File, fileHeaders multipart.FileHeader) (*minio.UploadInfo, *uuid.UUID, error) {
	defer file.Close()
	bucket_uuid := uuid.New()
	err := minioCreateBucket(bucket_uuid)
	if err != nil {
		return nil, nil, err
	}
	upload_info, err := minioClient.PutObject(
		context.Background(), //	Context (would it be better to use gin.Context ?)
		bucket_uuid.String(), //	Bucket name (use uuid)
		fileHeaders.Filename, //	File name (need potential system for this)
		file,                 //	Actual file
		fileHeaders.Size,     //	File size (for minio to optimize the upload, will need to inspect how it handles large files)
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	return &upload_info, &bucket_uuid, err
}

func minioPushFileNamed(bucket_uuid uuid.UUID, file multipart.File, fileHeaders multipart.FileHeader) (*minio.UploadInfo, error) {
	defer file.Close()
	_, err := minioClient.StatObject(context.Background(), bucket_uuid.String(), fileHeaders.Filename, minio.GetObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchBucket" || errResponse.Code == "InvalidBucketName" {
			return nil, errors.New("Invalid bucket.")
		}
		if errResponse.Code == "NoSuchKey" {
			upload_info, err := minioClient.PutObject(
				context.Background(),
				bucket_uuid.String(),
				fileHeaders.Filename,
				file,
				fileHeaders.Size,
				minio.PutObjectOptions{ContentType: "application/octet-stream"},
			)
			return &upload_info, err
		}
	}
	return nil, errors.New("File of given name already exists.")
}

func minioPutFileNamed(bucket_uuid uuid.UUID, file multipart.File, fileHeaders multipart.FileHeader) (*minio.UploadInfo, error) {
	defer file.Close()
	upload_info, err := minioClient.PutObject(
		context.Background(),
		bucket_uuid.String(),
		fileHeaders.Filename,
		file,
		fileHeaders.Size,
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	return &upload_info, err
}

func minioGetItemNamed(bucket_uuid uuid.UUID, item_name string) (*minio.Object, error) {
	return minioClient.GetObject(context.Background(), bucket_uuid.String(), item_name, minio.GetObjectOptions{})
}

func minioDeleteBucket(bucket_uuid uuid.UUID) error {
	objects := minioClient.ListObjects(context.Background(), bucket_uuid.String(), minio.ListObjectsOptions{})
	minioClient.RemoveObjects(context.Background(), bucket_uuid.String(), objects, minio.RemoveObjectsOptions{})
	return minioClient.RemoveBucket(context.Background(), bucket_uuid.String())
}

func minioCreateBucket(bucket_uuid uuid.UUID) error {
	return minioClient.MakeBucket(context.Background(), bucket_uuid.String(), minio.MakeBucketOptions{})
}

func minioListBuckets() ([]minio.BucketInfo, error) {
	return minioClient.ListBuckets(context.Background())
}
