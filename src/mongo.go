package main

import (
	"context"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProcessDefinition struct {
	ID           uuid.UUID `bson:"_id,omitempty"`
	Stage        int       `bson:"stage"`
	Status       string    `bson:"status"`
	LastFilename string    `bson:"lastfilename"`
	Routing      []Queue   `bson:"process"`
}

var mongoDB = "Minio"
var mongoCol = "BucketDetails"
var statWait = "waiting"
var statWork = "processing"
var statEnd = "Completed"

func mongoAddDetails(data ProcessDefinition) (*mongo.InsertOneResult, error) {
	data.Stage = 0
	data.Status = statWait
	collection := mongoClient.Database(mongoDB).Collection(mongoCol)
	return collection.InsertOne(context.Background(), data)
}

func mongoUpdateStatus(id uuid.UUID, status string) error {
	collection := mongoClient.Database(mongoDB).Collection(mongoCol)
	update := collection.FindOneAndUpdate(context.Background(), bson.M{"_id": id}, bson.M{"$set": map[string]interface{}{"status": status}})
	return update.Err()
}

func mongoUpdateStage(id uuid.UUID, stage int) error {
	collection := mongoClient.Database(mongoDB).Collection(mongoCol)
	update := collection.FindOneAndUpdate(context.Background(), bson.M{"_id": id}, bson.M{"$set": map[string]interface{}{"stage": stage, "status": "waiting"}})
	return update.Err()
}

func mongoUpdateLast(id uuid.UUID, last string) error {
	collection := mongoClient.Database(mongoDB).Collection(mongoCol)
	update := collection.FindOneAndUpdate(context.Background(), bson.M{"_id": id}, bson.M{"$set": map[string]interface{}{"lastfilename": last}})
	return update.Err()
}

func mongoUpdateDetails(id uuid.UUID, obj map[string]interface{}) error {
	collection := mongoClient.Database(mongoDB).Collection(mongoCol)
	update := collection.FindOneAndUpdate(context.Background(), bson.M{"_id": id}, bson.M{"$set": obj})
	return update.Err()
}

func mongoNextStage(id uuid.UUID) error {
	collection := mongoClient.Database(mongoDB).Collection(mongoCol)
	update := collection.FindOneAndUpdate(context.Background(), bson.M{"_id": id}, bson.M{"$inc": map[string]interface{}{"stage": 1}})
	if update.Err() != nil {
		return update.Err()
	}
	update = collection.FindOneAndUpdate(context.Background(), bson.M{"_id": id}, bson.M{"$set": map[string]interface{}{"status": "waiting"}})
	return update.Err()
}

func mongoGetNextQue(id uuid.UUID) (*Queue, error) {
	item, err := mongoGetDetails(id)
	if err != nil {
		return nil, err
	}
	if item.Stage+1 < len(item.Routing) {
		return &item.Routing[item.Stage+1], nil
	}
	return nil, nil
}

func mongoGetDetails(id uuid.UUID) (*ProcessDefinition, error) {
	collection := mongoClient.Database(mongoDB).Collection(mongoCol)
	var item ProcessDefinition
	err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}
