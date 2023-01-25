package main

import (
	"fmt"
	"runtime"

	"github.com/gin-gonic/gin" // gin for rest api
)

func main() {
	router := defaultRouter()

	//	0.0.0.0 is default docker use
	if os := runtime.GOOS; os == "windows" {
		fmt.Println(os)
		router.Run("localhost:8080")
	} else {
		fmt.Println(os)
		router.Run("0.0.0.0:8080")
	}

}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/api/")
	return router
}

func defaultRouter() *gin.Engine {
	router := gin.Default()
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

	return router
}
