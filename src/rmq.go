package main

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Returns minio client with default or cmd flag variables

type RoutingMap struct {
	Src string
	Dst string
}

// Returns amqp.Connection for RabbitMQ credentials
func getRMQConnection() (*amqp.Connection, error) {
	return amqp.Dial("amqp://" + rmq_cred_id + ":" + rmq_cred_key + "@" + rmq_host + ":" + strconv.Itoa(rmq_port) + "/")
}

func rmqBasicPublish(info RoutingMap, que_name string) (*string, error) {
	// parse json to bytes
	body_bytes, err := json.Marshal(info)
	if err != nil {
		return nil, errors.New("Failed to marshal json RoutingMap provided.")
	}
	// get RMQ connection
	rmq, err := getRMQConnection()
	if err != nil {
		return nil, errors.New("Failed to establish RMQ connection.")
	}
	//	Get channel
	ch, err := rmq.Channel()
	if err != nil {
		return nil, errors.New("Failed to create RMQ channel.")
	}
	// declare que
	q, err := ch.QueueDeclare(
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
	err = ch.PublishWithContext(
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
