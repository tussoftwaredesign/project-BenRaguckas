package main

import (
	"errors"
	"flag"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

// gin alias (for extending method) (Using Embedding as type definitions did not work as expected)
type GinContext struct{ *gin.Context }

// single file upload identifier
var form_file_identifier = "file"

// Connection properties
// var minio_host, rmq_host string
// var minio_port, rmq_port int
var minio_serv, rmq_serv string

// Credential properties
var minio_cred_id, minio_cred_key string
var rmq_cred_id, rmq_cred_key string

// Initialize cmd flags
func init() {
	//	NEW
	flag.StringVar(&minio_serv, "minio_serv", "min.localdev.me", "Minio host ip to which to communicate with.")
	flag.StringVar(&rmq_serv, "rmq_serv", "localhost:5672", "Minio host ip to which to communicate with.")
	//	Minio credentials
	flag.StringVar(&minio_cred_id, "minio_cred_id", "minio", "Minio access id (username).")
	flag.StringVar(&minio_cred_key, "minio_cred_key", "minio123", "Minio access key (password).")
	//	RabbitMQ credentials
	flag.StringVar(&rmq_cred_id, "rmq_cred_id", "guest", "RabbitMQ access id (username).")
	flag.StringVar(&rmq_cred_key, "rmq_cred_key", "guest", "RabbitMQ access key(password).")

	flag.Parse()
	//	Debug print
	fmt.Printf("MINIO: %s @ %s:%s\n", minio_serv, minio_cred_id, minio_cred_key)
	fmt.Printf("RMQ: %s @ %s:%s\n", rmq_serv, rmq_cred_id, rmq_cred_key)
}

func listItems(c *gin.Context, opts MethodOptions) error {
	if opts.PathParams == nil || opts.PathParams[0] == "" {
		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
	}
	objects, err := getMinioBucketItems(opts.PathParams[0])
	if err != nil {
		return err
	}
	//	Return nil error and response string with details
	object_list := []string{}
	for obj := range objects {
		if obj.Err != nil {
			fmt.Println(obj.Err)
			return err
		}
		object_list = append(object_list, obj.Key)
	}
	c.String(http.StatusOK, fmt.Sprintf("%s", object_list))
	return nil
}

func getItem(c *gin.Context, opts MethodOptions) error {
	if opts.PathParams == nil || len(opts.PathParams) < 1 || opts.PathParams[0] == "" {
		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
	}
	if len(opts.PathParams) < 2 || opts.PathParams[1] == "" {
		return errors.New("MethodOptions.PathParams[1] empty or nil and is required.")
	}
	obj, err := getMinioBucketItem(opts.PathParams[0], opts.PathParams[1])
	if err != nil {
		return err
	}
	//	Exclusive error response (This does not fail.. ever...)
	if err != nil {
		return err
	}
	stat, err := obj.Stat()
	//	Exclusive error response
	if err != nil {
		return err
	}
	defer obj.Close()
	c.DataFromReader(http.StatusOK, stat.Size, "application/octet-stream", obj, map[string]string{"Content-Disposition": `'attachment; filename="image.png"`})
	return nil
}

func pushItem(c *gin.Context, opts MethodOptions) error {
	//	Check for file identifier
	if opts.FileIdents == nil || len(opts.FileIdents) < 1 || opts.FileIdents[0] == "" {
		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
	}
	file, fileheaders := getFormFile(c, opts.FileIdents[0])
	if file == nil {
		return errors.New("Unable to read file. No file included.")
	}
	info, err := pushMinioFile(file, fileheaders)
	if err != nil {
		return err
	}
	c.String(http.StatusOK, fmt.Sprintf("File uploaded: \n\tbucket: %s\n\tkey: %s\n\tsize: %d", info.Bucket, info.Key, info.Size))
	return nil
}

func putItem(c *gin.Context, opts MethodOptions) error {
	//	Check for bucket_uuid
	if opts.PathParams == nil || len(opts.PathParams) < 1 || opts.PathParams[0] == "" {
		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
	}
	//	Check for file identifier
	if opts.FileIdents == nil || len(opts.FileIdents) < 1 || opts.FileIdents[0] == "" {
		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
	}
	file, fileheaders := getFormFile(c, opts.FileIdents[0])
	if file == nil {
		return errors.New("Unable to read file. No file included.")
	}
	info, err := putMinioFile(file, fileheaders, opts.PathParams[0])
	if err != nil {
		return err
	}
	c.String(http.StatusOK, fmt.Sprintf("File updated: \n\tbucket: %s\n\tkey: %s\n\tsize: %d", info.Bucket, info.Key, info.Size))
	return nil
}

func putItemNamed(c *gin.Context, opts MethodOptions) error {
	//	Check for bucket_uuid
	if opts.PathParams == nil || len(opts.PathParams) < 1 || opts.PathParams[0] == "" {
		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
	}
	//	Check for item_name
	if opts.PathParams == nil || len(opts.PathParams) < 2 || opts.PathParams[1] == "" {
		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
	}
	//	Check for file identifier
	if opts.FileIdents == nil || len(opts.FileIdents) < 1 || opts.FileIdents[0] == "" {
		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
	}
	file, fileheaders := getFormFile(c, opts.FileIdents[0])
	if file == nil {
		return errors.New("Unable to read file. No file included.")
	}
	info, err := putMinioFileNamed(file, fileheaders, opts.PathParams[0], opts.PathParams[1])
	if err != nil {
		return err
	}
	c.String(http.StatusOK, fmt.Sprintf("File updated: \n\tbucket: %s\n\tkey: %s\n\tsize: %d", info.Bucket, info.Key, info.Size))
	return nil
}

// HARD copy_paste + more of putItem
func processPredefinedBackground(c *gin.Context, opts MethodOptions) error {
	//	Check for file identifier
	if opts.FileIdents == nil || len(opts.FileIdents) < 1 || opts.FileIdents[0] == "" {
		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
	}
	file, fileheaders := getFormFile(c, opts.FileIdents[0])
	if file == nil {
		return errors.New("Unable to read file. No file included.")
	}
	info, err := pushMinioFile(file, fileheaders)
	if err != nil {
		return err
	}
	data := RoutingMap{fmt.Sprintf("/custom/%s/%s", info.Bucket, info.Key), fmt.Sprintf("/custom/%s", info.Bucket)}
	routing_key, err := rmqBasicPublish(data, "background_queue")
	if err != nil {
		return err
	}
	c.String(http.StatusOK, fmt.Sprintf("File put up for processing: \n\tbucket: %s\n\tkey: %s\n\tsize: %d\n\trouting: %s", info.Bucket, info.Key, info.Size, *routing_key))
	return nil
}

// HARD copy_paste + more of putItem
func processPredefinedGray(c *gin.Context, opts MethodOptions) error {
	//	Check for file identifier
	if opts.FileIdents == nil || len(opts.FileIdents) < 1 || opts.FileIdents[0] == "" {
		return errors.New("MethodOptions.PathParams[0] empty or nil and is required.")
	}
	file, fileheaders := getFormFile(c, opts.FileIdents[0])
	if file == nil {
		return errors.New("Unable to read file. No file included.")
	}
	info, err := pushMinioFile(file, fileheaders)
	if err != nil {
		return err
	}
	data := RoutingMap{fmt.Sprintf("/custom/%s/%s", info.Bucket, info.Key), fmt.Sprintf("/custom/%s", info.Bucket)}
	routing_key, err := rmqBasicPublish(data, "gray_queue")
	if err != nil {
		return err
	}
	c.String(http.StatusOK, fmt.Sprintf("File put up for processing: \n\tbucket: %s\n\tkey: %s\n\tsize: %d\n\trouting: %s", info.Bucket, info.Key, info.Size, *routing_key))
	return nil
}

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
