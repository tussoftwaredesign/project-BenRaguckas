package main

import (
	"fmt"
	"runtime"

	"github.com/gin-gonic/gin" // gin for rest api
)

var methods = map[string]func(*gin.Context){}

// Idea to try using alias for extending gin engime method and mapping to base function (router.POST , router.GET ,...)
type GinEngine struct{ *gin.Engine }

func (r GinEngine) MapApi(api CustomEndpoint) {

}

var config ApiConfig
var router *gin.Engine

func init() {
	config = parseConfig()
	router = gin.Default()
}

func main() {

	//	Test
	for _, api := range config.Apis {
		switch api.Uri.Method {
		case "GET":
			router.GET(api.Uri.Value, MainAction)
		case "POST":
			router.POST(api.Uri.Value, MainAction)
		case "PUT":
			router.PUT(api.Uri.Value, MainAction)
		case "DELETE":
			router.DELETE(api.Uri.Value, MainAction)
		}
	}

	//	Comparison request from graveyard
	// router.GET("/preset", func(ctx *gin.Context) {
	// 	getMinioFile(ctx, "244d6713-1686-4974-82ec-a0ddb29e53bf", "image.jpg")
	// })

	// router := defaultRouter()

	//	0.0.0.0 is default docker use
	if os := runtime.GOOS; os == "windows" {
		fmt.Println(os)
		router.Run("localhost:8080")
	} else {
		fmt.Println(os)
		router.Run("0.0.0.0:8080")
	}

}

func defaultRouter() *gin.Engine {
	router := gin.Default()

	//	Actual endpoints
	router.POST("/api/upload", addNewMinioFile)
	router.PUT("/api/upload/:x", func(c *gin.Context) {
		uuid := c.Param("x")
		putMinioFile2(c, uuid)
	})
	router.GET("api/get/:x/:y", func(c *gin.Context) {
		bucket := c.Param("x")
		file_name := c.Param("y")
		getMinioFile(c, bucket, file_name)
	})

	//	Test endpoints
	router.GET("/test", testDefault)
	router.GET("/test/minio", testMinioHealth)

	router.MaxMultipartMemory = 8 << 20

	router.POST("/test/upload", func(c *gin.Context) {
		testSaveFile(c, "file")
	})

	router.GET("/test/minio/create/:name", func(c *gin.Context) {
		name := c.Param("name")
		testMinioCreateBucket(c, name)
	})

	router.POST("/test/minio/upload", func(c *gin.Context) {
		testMinioAddFile(c, "file", "test-bucketname")
	})

	router.POST("/test/rmq/publish", func(c *gin.Context) {
		body := c.Request.FormValue("message")
		testMQPublish(c, body)
	})
	return router
}
