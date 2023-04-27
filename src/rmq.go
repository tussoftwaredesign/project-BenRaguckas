package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Returns minio client with default or cmd flag variables

type BackendTask struct {
	Src    string
	Dst    string
	Stat   string
	Params *json.RawMessage
}

// var rmq *amqp.Connection
var rmqConn *amqp.Connection
var rmqChannel *amqp.Channel

var defApiBucketIdentifier = ":bucket"
var defApiFnameIdentifier = ":fname"

func establishRMQConnection() {
	var err error
	rmqConn, err = amqp.Dial("amqp://" + rmq_cred_id + ":" + rmq_cred_key + "@" + rmq_serv + "/")
	if err != nil {
		fmt.Printf("Failed at creating RabbitMQ client using: %s @ %s:%s\n", rmq_serv, rmq_cred_id, rmq_cred_key)
	}
	rmqChannel, err = rmqConn.Channel()
	if err != nil {
		println("Failed to create rmq channel.")
		os.Exit(4)
	}
}

func rmqBasicPublish(info BackendTask, que_name string) (*string, error) {
	// Check connection and reconnect in case
	if rmqConn.IsClosed() {
		establishRMQConnection()
	}
	// parse json to bytes
	body_bytes, err := json.Marshal(info)
	if err != nil {
		return nil, errors.New("Failed to marshal json RoutingMap provided.")
	}
	// declare que
	q, err := rmqChannel.QueueDeclare(
		que_name, // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return nil, errors.New("Failed to create RMQ queue.")
	}
	// create timeout context (in the case of heavy loads, may be redundant for orchestated use)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//	Publish the message info body
	err = rmqChannel.PublishWithContext(
		ctx,
		"",     // Exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/json",
			Body:        body_bytes,
		},
	)
	if err != nil {
		return nil, errors.New("Failed to publish RMQ message.")
	}
	return &q.Name, nil
}

func rmqBodyBuild(bucket_uuid uuid.UUID, que Queue) BackendTask {
	// if que has input specified
	var src_url string
	if que.Input != nil && *que.Input != "" {
		src_url = strings.Replace(config.Def_apis.GetItemNamed, defApiBucketIdentifier, bucket_uuid.String(), -1)
		src_url = strings.Replace(src_url, defApiFnameIdentifier, *que.Input, -1)
	} else {
		src_url = strings.Replace(config.Def_apis.GetItem, defApiBucketIdentifier, bucket_uuid.String(), -1)
	}
	// if que has output specified
	var dst_url string
	if que.Output != nil && *que.Output != "" {
		dst_url = strings.Replace(config.Def_apis.PutItemNamed, defApiBucketIdentifier, bucket_uuid.String(), -1)
		dst_url = strings.Replace(dst_url, defApiFnameIdentifier, *que.Output, -1)
	} else {
		dst_url = strings.Replace(config.Def_apis.PutItem, defApiBucketIdentifier, bucket_uuid.String(), -1)
	}
	// Construct object
	body := BackendTask{
		Src:  src_url,
		Dst:  dst_url,
		Stat: strings.Replace(config.Def_apis.PostStatus, defApiBucketIdentifier, bucket_uuid.String(), -1),
	}
	// If Params are present
	// if que.Params != nil && *que.Params != "" {
	if que.Params != nil {
		body.Params = que.Params
	}
	return body
}
