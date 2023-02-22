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
var minio_host, rmq_host string
var minio_port, rmq_port int

// Credential properties
var minio_cred_id, minio_cred_key string
var rmq_cred_id, rmq_cred_key string

// Initialize cmd flags
func init() {
	//	endpoints (variable name, default, description)
	flag.StringVar(&minio_host, "minio_host", "localhost", "Minio host ip to which to communicate with.")
	flag.IntVar(&minio_port, "minio_port", 30036, "Minio host port.")
	flag.StringVar(&rmq_host, "rmq_host", "localhost", "RabbitMQ host ip to which to communicate with.")
	flag.IntVar(&rmq_port, "rmq_port", 30034, "RabbitMQ host port.")
	//	Minio credentials
	flag.StringVar(&minio_cred_id, "minio_cred_id", "TE36SvMRgIWxe8lP", "Minio access id (username).")
	flag.StringVar(&minio_cred_key, "minio_cred_key", "ncYM3dcwASrMNMk1Y7AQZAeHvA2SuooZ", "Minio access key (password).")
	//	RabbitMQ credentials
	flag.StringVar(&rmq_cred_id, "rmq_cred_id", "guest", "RabbitMQ access id (username).")
	flag.StringVar(&rmq_cred_key, "rmq_cred_key", "guest", "RabbitMQ access key(password).")

	flag.Parse()
	fmt.Printf("MINIO: %s:%d\n", minio_host, minio_port)
	fmt.Printf("MINIO CREDS: %s:%s\n", minio_cred_id, minio_cred_key)
	fmt.Printf("RMQ: %s:%d\n", rmq_host, rmq_port)
	fmt.Printf("RMQ CREDS: %s:%s\n", rmq_cred_id, rmq_cred_key)
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
