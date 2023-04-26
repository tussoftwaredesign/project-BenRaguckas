package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin" // gin for rest api
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var methods = map[string]func(*gin.Context){}

// Idea to try using alias for extending gin engime method and mapping to base function (router.POST , router.GET ,...)
type GinEngine struct{ *gin.Engine }

func (r GinEngine) MapApi(api CustomEndpoint) {

}

var config ApiConfig
var max_attempts = 5

var mongoClient *mongo.Client

func init() {
	config = parseConfig()
}

func main() {
	//	Establish minio client
	var err error
	minioClient, err = minio.New(minio_serv, &minio.Options{
		Creds:  credentials.NewStaticV4(minio_cred_id, minio_cred_key, ""),
		Secure: false,
	})
	if err != nil {
		fmt.Printf("Failed at creating Minio client using: %s @ %s:%s\n", minio_serv, minio_cred_id, minio_cred_key)
		os.Exit(3)
	}

	//	Establish mongo client
	// mongoURI := "mongodb://" + mongo_cred_id + ":" + mongo_cred_key + "@" + mongo_serv
	mongoURI := "mongodb://" + mongo_serv
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err = mongo.Connect(context.Background(), opts)
	if err != nil {
		fmt.Printf("Failed at creating Mongo client using: %s @ %s:%s\n", mongo_serv, mongo_cred_id, mongo_cred_key)
		os.Exit(3)
	}

	//	Test connections
	connections := false
	attempt := 0
	rmqConnnection := false
	for !connections && attempt < max_attempts {
		attempt++
		connections = true

		// Test minio connection
		_, err = minioClient.ListBuckets(context.Background())
		if err != nil {
			connections = false
			fmt.Printf("%d / %d Failed connecting to Minio: %s @ %s:%s\n", attempt, max_attempts, minio_serv, minio_cred_id, minio_cred_key)
		}

		//	Establish rabbitMQ client
		if !rmqConnnection {
			// Bootleg fix
			establishRMQConnection()
			rmqConnnection = true
			// rmq, err := amqp.Dial("amqp://" + rmq_cred_id + ":" + rmq_cred_key + "@" + rmq_serv + "/")
			// rmqConnnection = true
			// if err != nil {
			// 	rmqConnnection = false
			// 	connections = false
			// 	fmt.Printf("%d / %d Failed at creating RabbitMQ client using: %s @ %s:%s\n", attempt, max_attempts, rmq_serv, rmq_cred_id, rmq_cred_key)
			// }
			// rmqChannel, err = rmq.Channel()
			// if err != nil {
			// 	println("Failed to create rmq channel.")
			// 	os.Exit(4)
			// }
		}

		// Test mongo connection
		err = mongoClient.Ping(context.Background(), nil)
		if err != nil {
			connections = false
			fmt.Printf("%d / %d Failed connecting to MongoDB: %s @ %s:%s\n", attempt, max_attempts, mongo_serv, mongo_cred_id, mongo_cred_key)
		}
	}
	if !connections {
		println("Exiting due to connection failures.")
		os.Exit(4)
	}

	routerHandles()
}

func routerHandles() {
	// Router
	router := gin.Default()

	//	Default Routing
	// Basic put
	router.Handle("PUT", config.Def_apis.PutItem, func(c *gin.Context) {
		err := defaultPutItem(c, "bucket", "file")
		if err != nil {
			c.Status(http.StatusInternalServerError)
		}
		c.Status(http.StatusOK)
	})
	// Basic put using file name specified in URL
	router.Handle("PUT", config.Def_apis.PutItemNamed, func(c *gin.Context) {
		err := defaultPutItemNamed(c, "bucket", "fname", "file")
		if err != nil {
			c.Status(http.StatusInternalServerError)
		}
		c.Status(http.StatusOK)
	})
	//	Get default item of bucket
	router.Handle("GET", config.Def_apis.GetItem, func(c *gin.Context) {
		err := defaultGetItem(c, "bucket")
		if err != nil {
			c.Status(http.StatusInternalServerError)
		}
	})
	//	Get specific item of bucket (based on URL)
	router.Handle("GET", config.Def_apis.GetItemNamed, func(c *gin.Context) {
		err := defaultGetItemNamed(c, "bucket", "fname")
		if err != nil {
			c.Status(http.StatusInternalServerError)
		}
	})
	// Post for status updates
	router.Handle("POST", config.Def_apis.PostStatus, func(c *gin.Context) {
		err := defaultPostStatus(c, "bucket", "status")
		if err != nil {
			c.Status(http.StatusInternalServerError)
		}
		c.Status(http.StatusOK)
	})

	// Custom Routings
	for _, endpoint := range config.EndPoints {
		// If simple endpoints handle them
		e := endpoint
		if len(endpoint.SimpleFunction) > 0 {
			for _, sfunc := range e.SimpleFunction {
				sf := sfunc
				uri := e.Uri
				router.Handle(sf.Method, e.Uri.Value, func(c *gin.Context) {
					opts := getSFuncOptions(c, uri, sf)
					func_return, err := Call(sf.FunctionName, c, opts)
					if func_return != nil {
						c.String(http.StatusInternalServerError, "Unexpected return value.")
					}
					if err != nil {
						c.String(http.StatusInternalServerError, "Error calling a method.")
					}
				})
				// router.Handle(action.Method, endpoint.Uri.Value, MainAction2)
			}
		} else if endpoint.DefinedRouting != nil {
			router.POST(endpoint.Uri.Value, func(c *gin.Context) {
				// restPostDefinedRouting(c, *testvar.DefinedRouting)
				restTest(c, *e.DefinedRouting)
			})
		} else { //	If not simple or custom do userDefined
			router.POST(endpoint.Uri.Value, func(c *gin.Context) {
				fmt.Println("USER ROUTING")
			})
		}

	}

	//	Start REST API
	router.Run(":8080")
}
