package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// gin alias (for extending method) (Using Embedding as type definitions did not work as expected)
type GinContext struct{ *gin.Context }

// single file upload identifier
var form_file_identifier = "file"

// Connection services
var minio_serv, rmq_serv, mongo_serv string

// Credentials
var minio_cred_id, minio_cred_key string
var rmq_cred_id, rmq_cred_key string
var mongo_cred_id, mongo_cred_key string

// Initialize cmd flags
func init() {
	//	Services
	flag.StringVar(&minio_serv, "minio_serv", "localhost:9000", "Minio address to which to communicate with.")
	flag.StringVar(&rmq_serv, "rmq_serv", "localhost:5672", "RabbitMQ address to which to communicate with.")
	flag.StringVar(&mongo_serv, "mongo_serv", "localhost:27017", "MongoDB address to which to communicate with.")
	//	Credentials
	flag.StringVar(&minio_cred_id, "minio_cred_id", "minio", "Minio access id (username).")
	flag.StringVar(&minio_cred_key, "minio_cred_key", "minio123", "Minio access key (password).")
	flag.StringVar(&rmq_cred_id, "rmq_cred_id", "guest", "RabbitMQ access id (username).")
	flag.StringVar(&rmq_cred_key, "rmq_cred_key", "guest", "RabbitMQ access key(password).")
	flag.StringVar(&mongo_cred_id, "mongo_cred_id", "mongo", "MongoDB access id (username).")
	flag.StringVar(&mongo_cred_key, "mongo_cred_key", "mongo123", "MongoDB access key(password).")

	flag.Parse()
	//	Debug print
	fmt.Printf("MINIO: %s @ %s:%s\n", minio_serv, minio_cred_id, minio_cred_key)
	fmt.Printf("RMQ: %s @ %s:%s\n", rmq_serv, rmq_cred_id, rmq_cred_key)
	fmt.Printf("MONGO: %s @ %s:%s\n", mongo_serv, mongo_cred_id, mongo_cred_key)
}

// func listItems(c *gin.Context, opts MethodOptions) error {
// 	if opts.PathParams == nil || opts.PathParams[0] == "" {
// 		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
// 	}
// 	objects, err := getMinioBucketItems(opts.PathParams[0])
// 	if err != nil {
// 		return err
// 	}
// 	//	Return nil error and response string with details
// 	object_list := []string{}
// 	for obj := range objects {
// 		if obj.Err != nil {
// 			fmt.Println(obj.Err)
// 			return err
// 		}
// 		object_list = append(object_list, obj.Key)
// 	}
// 	c.String(http.StatusOK, fmt.Sprintf("%s", object_list))
// 	return nil
// }

// func getItem(c *gin.Context, opts MethodOptions) error {
// 	if opts.PathParams == nil || len(opts.PathParams) < 1 || opts.PathParams[0] == "" {
// 		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
// 	}
// 	if len(opts.PathParams) < 2 || opts.PathParams[1] == "" {
// 		return errors.New("MethodOptions.PathParams[1] empty or nil and is required.")
// 	}
// 	obj, err := getMinioBucketItem(opts.PathParams[0], opts.PathParams[1])
// 	if err != nil {
// 		return err
// 	}
// 	//	Exclusive error response (This does not fail.. ever...)
// 	if err != nil {
// 		return err
// 	}
// 	stat, err := obj.Stat()
// 	//	Exclusive error response
// 	if err != nil {
// 		return err
// 	}
// 	defer obj.Close()
// 	c.DataFromReader(http.StatusOK, stat.Size, "application/octet-stream", obj, map[string]string{"Content-Disposition": `'attachment; filename="image.png"`})
// 	return nil
// }

// func pushItem(c *gin.Context, opts MethodOptions) error {
// 	//	Check for file identifier
// 	if opts.FileIdents == nil || len(opts.FileIdents) < 1 || opts.FileIdents[0] == "" {
// 		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
// 	}
// 	file, fileheaders := getFormFile(c, opts.FileIdents[0])
// 	if file == nil {
// 		return errors.New("Unable to read file. No file included.")
// 	}
// 	info, _, err := pushMinioFile(file, fileheaders)
// 	if err != nil {
// 		return err
// 	}
// 	c.String(http.StatusOK, fmt.Sprintf("File uploaded: \n\tbucket: %s\n\tkey: %s\n\tsize: %d", info.Bucket, info.Key, info.Size))
// 	return nil
// }

// func putItem(c *gin.Context, opts MethodOptions) error {
// 	//	Check for bucket_uuid
// 	if opts.PathParams == nil || len(opts.PathParams) < 1 || opts.PathParams[0] == "" {
// 		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
// 	}
// 	//	Check for file identifier
// 	if opts.FileIdents == nil || len(opts.FileIdents) < 1 || opts.FileIdents[0] == "" {
// 		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
// 	}
// 	file, fileheaders := getFormFile(c, opts.FileIdents[0])
// 	if file == nil {
// 		return errors.New("Unable to read file. No file included.")
// 	}
// 	info, err := putMinioFile(file, fileheaders, opts.PathParams[0])
// 	if err != nil {
// 		return err
// 	}
// 	c.String(http.StatusOK, fmt.Sprintf("File updated: \n\tbucket: %s\n\tkey: %s\n\tsize: %d", info.Bucket, info.Key, info.Size))
// 	return nil
// }

// func putItemNamed(c *gin.Context, opts MethodOptions) error {
// 	//	Check for bucket_uuid
// 	if opts.PathParams == nil || len(opts.PathParams) < 1 || opts.PathParams[0] == "" {
// 		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
// 	}
// 	//	Check for item_name
// 	if opts.PathParams == nil || len(opts.PathParams) < 2 || opts.PathParams[1] == "" {
// 		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
// 	}
// 	//	Check for file identifier
// 	if opts.FileIdents == nil || len(opts.FileIdents) < 1 || opts.FileIdents[0] == "" {
// 		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
// 	}
// 	file, fileheaders := getFormFile(c, opts.FileIdents[0])
// 	if file == nil {
// 		return errors.New("Unable to read file. No file included.")
// 	}
// 	info, err := putMinioFileNamed(file, fileheaders, opts.PathParams[0], opts.PathParams[1])
// 	if err != nil {
// 		return err
// 	}
// 	c.String(http.StatusOK, fmt.Sprintf("File updated: \n\tbucket: %s\n\tkey: %s\n\tsize: %d", info.Bucket, info.Key, info.Size))
// 	return nil
// }

// HARD copy_paste + more of putItem
// func processPredefinedBackground(c *gin.Context, opts MethodOptions) error {
// 	//	Check for file identifier
// 	if opts.FileIdents == nil || len(opts.FileIdents) < 1 || opts.FileIdents[0] == "" {
// 		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
// 	}
// 	file, fileheaders := getFormFile(c, opts.FileIdents[0])
// 	if file == nil {
// 		return errors.New("Unable to read file. No file included.")
// 	}
// 	info, err := pushMinioFile(file, fileheaders)
// 	if err != nil {
// 		return err
// 	}
// 	data := BackendTask{fmt.Sprintf("/custom/%s/%s", info.Bucket, info.Key), fmt.Sprintf("/custom/%s", info.Bucket), ""}
// 	routing_key, err := rmqBasicPublish(data, "background_queue")
// 	if err != nil {
// 		return err
// 	}
// 	c.String(http.StatusOK, fmt.Sprintf("File put up for processing: \n\tbucket: %s\n\tkey: %s\n\tsize: %d\n\trouting: %s", info.Bucket, info.Key, info.Size, *routing_key))
// 	return nil
// }

// // HARD copy_paste + more of putItem
// func processPredefinedGray(c *gin.Context, opts MethodOptions) error {
// 	//	Check for file identifier
// 	if opts.FileIdents == nil || len(opts.FileIdents) < 1 || opts.FileIdents[0] == "" {
// 		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
// 	}
// 	file, fileheaders := getFormFile(c, opts.FileIdents[0])
// 	if file == nil {
// 		return errors.New("Unable to read file. No file included.")
// 	}
// 	info, err := pushMinioFile(file, fileheaders)
// 	if err != nil {
// 		return err
// 	}
// 	data := BackendTask{fmt.Sprintf("/custom/%s/%s", info.Bucket, info.Key), fmt.Sprintf("/custom/%s", info.Bucket), ""}
// 	routing_key, err := rmqBasicPublish(data, "gray_queue")
// 	if err != nil {
// 		return err
// 	}
// 	c.String(http.StatusOK, fmt.Sprintf("File put up for processing: \n\tbucket: %s\n\tkey: %s\n\tsize: %d\n\trouting: %s", info.Bucket, info.Key, info.Size, *routing_key))
// 	return nil
// }

func getFormFile(c *gin.Context, identifier string) (multipart.File, multipart.FileHeader) {
	fileHeaders, err := c.FormFile(identifier)
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

// The Content
func restPostDefinedRouting(c *gin.Context, rout DefinedRouting) error {
	//	Create uuid
	bucket_uuid := uuid.New()

	//	Create bucket
	err := minioClient.MakeBucket(context.Background(), bucket_uuid.String(), minio.MakeBucketOptions{})
	if err != nil {
		return err
	}

	//	Get file from body
	file, fileheaders := getFormFile(c, rout.ObjectIdent)
	if file == nil {
		return errors.New("Unable to read file. No file included.")
	}

	// Upload file to bucket
	upload_info, err := minioClient.PutObject(
		context.Background(), //	Context (would it be better to use gin.Context ?)
		bucket_uuid.String(), //	Bucket name (use uuid)
		fileheaders.Filename, //	File name (need potential system for this)
		file,                 //	Actual file
		fileheaders.Size,     //	File size (for minio to optimize the upload, will need to inspect how it handles large files)
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	if err != nil {
		return err
	}

	// Add mongoDB data
	mongoData := ProcessDefinition{
		ID: bucket_uuid,
		// Stage:    0,
		// Status:   "waiting",
		LastFilename: fileheaders.Filename,
		Routing:      rout.Ques,
	}
	//	Add data to mongo
	_, err = mongoAddDetails(mongoData)
	if err != nil {
		return err
	}
	// Route to rabbitMQ
	data := BackendTask{
		Src:    "SRC URL:" + bucket_uuid.String(),
		Dst:    "DST URL:" + bucket_uuid.String(),
		Params: rout.Ques[0].Params,
	}
	// {fmt.Sprintf("/custom/%s/%s", info.Bucket, info.Key), fmt.Sprintf("/custom/%s", info.Bucket), ""}
	routing_key, err := rmqBasicPublish(data, rout.Ques[0].Que)
	if err != nil {
		return err
	}
	// end
	_ = routing_key
	c.String(http.StatusOK, fmt.Sprintf("File put up for processing: \n\tbucket: %s\n\tkey: %s\n\tsize: %d", upload_info.Bucket, upload_info.Key, upload_info.Size))
	return nil
}

// Custom body
func restTest(c *gin.Context, rout DefinedRouting) error {
	// Get Form file
	file, file_headers := getFormFile(c, rout.ObjectIdent)
	if file == nil {
		return errors.New("Failed to read file.")
	}

	// Push file and get uuid
	upload_info, bucket_uuid, err := minioPushFile(file, file_headers)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return err
	}

	// Push info to mongo
	meta_data := ProcessDefinition{
		ID:           *bucket_uuid,
		LastFilename: file_headers.Filename,
		Routing:      rout.Ques,
	}
	_, err = mongoAddDetails(meta_data)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return err
	}

	// RabbitMQues
	body := rmqBodyBuild(*bucket_uuid, rout.Ques[0])
	_, err = rmqBasicPublish(body, rout.Ques[0].Que)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return err
	}
	c.String(http.StatusOK, fmt.Sprintf("File put up for processing: \n\tbucket: %s\n\tkey: %s\n\tsize: %d", upload_info.Bucket, upload_info.Key, upload_info.Size))
	return nil
}

func restPostComplexProcess() {

}

func restPutStageProcess() {

}

// Puts item and progresses que if available
func defaultPutItem(c *gin.Context, bucket_ident string, file_ident string) error {
	// Parse UUID
	bucket_uuid, err := uuid.Parse(c.Param(bucket_ident))
	if err != nil {
		return err
	}
	// Parse File
	file, fileheaders := getFormFile(c, file_ident)
	if file == nil {
		return errors.New("Unable to read file. No file included.")
	}
	// commot put
	return defaultPut(bucket_uuid, file, fileheaders)
}

// Puts item and progresses que if available utilizing name from uri
func defaultPutItemNamed(c *gin.Context, bucket_ident string, file_name_ident string, file_ident string) error {
	// Parse UUID and file_name
	file_name := c.Param(file_name_ident)
	bucket_uuid, err := uuid.Parse(c.Param(bucket_ident))
	if err != nil {
		return err
	}
	// Parse File
	file, fileheaders := getFormFile(c, file_ident)
	if file == nil {
		return errors.New("Unable to read file. No file included.")
	}
	fileheaders.Filename = file_name
	// commont put
	return defaultPut(bucket_uuid, file, fileheaders)
}

// Common Put
func defaultPut(bucket_uuid uuid.UUID, file multipart.File, fileheaders multipart.FileHeader) error {
	minioPutFileNamed(bucket_uuid, file, fileheaders)
	mongoUpdateLast(bucket_uuid, fileheaders.Filename)
	// Check following stage
	que, err := mongoGetNextQue(bucket_uuid)
	if err != nil {
		return err
	}
	if que == nil {
		mongoUpdateStatus(bucket_uuid, statEnd)
		return nil
	}
	// RMQDATA
	data := rmqBodyBuild(bucket_uuid, *que)
	_, err = rmqBasicPublish(data, que.Que)
	if err != nil {
		return err
	}
	mongoNextStage(bucket_uuid)

	return nil
}
func defaultGetItem(c *gin.Context, bucket_ident string) error {
	bucket_uuid, err := uuid.Parse(c.Param(bucket_ident))
	if err != nil {
		return err
	}
	info, err := mongoGetDetails(bucket_uuid)
	return defaultGet(c, bucket_uuid, info.LastFilename)
}
func defaultGetItemNamed(c *gin.Context, bucket_ident string, file_name_ident string) error {
	file_name := c.Param(file_name_ident)
	bucket_uuid, err := uuid.Parse(c.Param(bucket_ident))
	if err != nil {
		return err
	}
	return defaultGet(c, bucket_uuid, file_name)
}
func defaultGet(c *gin.Context, bucket_uuid uuid.UUID, file_name string) error {
	file, err := minioGetItemNamed(bucket_uuid, file_name)
	defer file.Close()
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchBucket" || errResponse.Code == "InvalidBucketName" {
			c.String(http.StatusBadRequest, "Unable to find specified bucket.")
		}
		if errResponse.Code == "NoSuchKey" {
			c.String(http.StatusBadRequest, "Unable to find specified file.")
		}
		return err
	}
	stat, err := file.Stat()
	//	Exclusive error response
	if err != nil {
		c.String(http.StatusBadRequest, "No such file.")
		return err
	}

	c.DataFromReader(http.StatusOK, stat.Size, "application/octet-stream", file, map[string]string{"Content-Disposition": `'attachment; filename="` + stat.Key + `"`})
	return nil
}
func defaultPostStatus(c *gin.Context, bucket_ident string, form_ident string) error {
	bucket_uuid, err := uuid.Parse(c.Param(bucket_ident))
	if err != nil {
		return err
	}
	stat := c.Request.FormValue(form_ident)
	if stat == "" {
		stat = statWork
	}
	return mongoUpdateStatus(bucket_uuid, stat)
}

// Simple functions (not really)
func simpPostItem(c *gin.Context, opts MethodOptions) {
	upload_info, _, err := minioPushFile((*opts.Files)[0], (*opts.FileHeaders)[0])
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("File put up for processing: \n\tbucket: %s\n\tkey: %s\n\tsize: %d", upload_info.Bucket, upload_info.Key, upload_info.Size))
}
func simpPutItem(c *gin.Context, opts MethodOptions) {
	bucket_id, err := uuid.Parse((*opts.PathParams)[0])
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	upload_info, err := minioPutFileNamed(bucket_id, (*opts.Files)[0], (*opts.FileHeaders)[0])
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("File put up for processing: \n\tbucket: %s\n\tkey: %s\n\tsize: %d", upload_info.Bucket, upload_info.Key, upload_info.Size))
}
func simpPutItemNamed(c *gin.Context, opts MethodOptions) {
	bucket_id, err := uuid.Parse((*opts.PathParams)[0])
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if opts.FileHeaders != nil {
		(*opts.FileHeaders)[0].Filename = (*opts.PathParams)[1]
	}
	upload_info, err := minioPutFileNamed(bucket_id, (*opts.Files)[0], (*opts.FileHeaders)[0])
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("File put up for processing: \n\tbucket: %s\n\tkey: %s\n\tsize: %d", upload_info.Bucket, upload_info.Key, upload_info.Size))
}
func simpListBuckets(c *gin.Context, opts MethodOptions) {
	info, err := minioListBuckets()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, info)
}
func simpListBucketDetais(c *gin.Context, opts MethodOptions) {
	bucket_id, err := uuid.Parse((*opts.PathParams)[0])
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	info, err := mongoGetDetails(bucket_id)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusOK, info)
}
func simpGetItem(c *gin.Context, opts MethodOptions) {
	bucket_id, err := uuid.Parse((*opts.PathParams)[0])
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	info, err := mongoGetDetails(bucket_id)
	defaultGet(c, bucket_id, info.LastFilename)
}
func simpGetItemNamed(c *gin.Context, opts MethodOptions) {
	bucket_id, err := uuid.Parse((*opts.PathParams)[0])
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	defaultGet(c, bucket_id, (*opts.PathParams)[1])
}
func simpDeleteBucket(c *gin.Context, opts MethodOptions) {
	bucket_id, err := uuid.Parse((*opts.PathParams)[0])
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	err = minioDeleteBucket(bucket_id)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}
