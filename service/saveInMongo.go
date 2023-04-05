package service

import (
	"context"
	"crawl-file/connectMongoDb"
	"crawl-file/model"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

func SaveFileToMongoDb(doc []mongo.WriteModel) error {
	bulkWrite, err := connectMongoDb.Collection.BulkWrite(context.Background(), doc)
	if err != nil {
		return err
	}
	log.Printf("Inserted %d documents", bulkWrite.InsertedCount)
	return nil
}

func UpdateDataMongodb(domain model.Domain) error {
	filter := bson.M{"domain": domain.DomainUrl}

	// Define the update to apply
	update := bson.M{"$set": bson.M{
		"title":       domain.Title,
		"description": domain.Description,
		"keywords":    domain.Keywords,
		"owner":       domain.Owner,
		"expires":     domain.Expires,
		"created":     domain.Created,
		"status":      model.StatusUpdated,
	}}

	// Update the first document matching the filter
	result, err := connectMongoDb.Collection.UpdateOne(context.Background(), filter, update)
	if domain.Description != "none" {
		fmt.Println(result)
	}
	if err != nil {
		return err
	}

	fmt.Printf("Matched %v documents and modified %v documents.\n", result.MatchedCount, result.ModifiedCount)

	return nil
}
