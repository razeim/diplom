package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBSet() *mongo.Client {

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://razeim05:KdFiTEyUtLiUD2lY@cluster0.wsffdps.mongodb.net/Music?retryWrites=true&w=majority&appName=Cluster0"))

	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Println("failed to connect mongodb")
		return nil
	}

	fmt.Print("Successfully connected to mognodb")

	return client
}

var Client *mongo.Client = DBSet()

func UserData(client *mongo.Client, collectionName string) *mongo.Collection {

	var collection *mongo.Collection = client.Database("Music").Collection(collectionName)
	return collection
}
func ProductData(client *mongo.Client, collectionName string) *mongo.Collection {

	var productCollection *mongo.Collection = client.Database("Music").Collection(collectionName)
	return productCollection
}
