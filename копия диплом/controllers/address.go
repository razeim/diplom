package controllers

import (
	"beep/models"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddPaymentDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("id")

		if userID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "invalid search index"})
			c.Abort()
			return
		}

		payment, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.IndentedJSON(500, "Internal server error")

		}

		var payments models.PaymentDetails

		payments.Address_ID = primitive.NewObjectID()
		if err := c.BindJSON(&payments); err != nil {
			c.IndentedJSON(http.StatusNotAcceptable, err.Error())
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		matchFilter := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: payment}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$payment_details"}}}}
		group := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$payment_details_id"}, {Key: "count", Value: bson.D{primitive.E{Key: "$sum", Value: 1}}}}}}
		pointCursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{matchFilter, unwind, group})
		if err != nil {
			c.IndentedJSON(500, "Internal server error")
		}

		var paymentInfo []bson.M

		if err := pointCursor.All(ctx, &paymentInfo); err != nil {
			panic(err)
		}

		var size int32
		for _, address_no := range paymentInfo {
			count := address_no["count"]
			size = count.(int32)
		}

		if size < 2 {
			filter := bson.D{primitive.E{Key: "_id", Value: payment}}
			update := bson.D{primitive.E{Key: "$push", Value: bson.D{primitive.E{Key: "_id", Value: payments}}}}
			_, err := UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				fmt.Println(err)
			}

		} else {
			c.IndentedJSON(400, "Not Allowed")

		}
		defer cancel()
		ctx.Done()
	}

}

func EditPaymentDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("id")

		if userID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "invalid "})
			c.Abort()

			return
		}
		userObjectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(500, "Internal Server error")
			return
		}
		var paymentDetails models.PaymentDetails
		if err := c.BindJSON(&paymentDetails); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		filter := bson.D{primitive.E{Key: "_id", Value: userObjectID}}
		update := bson.D{{Key: "$set", Value: bson.D{{Key: "payment_details.card_number", Value: paymentDetails.CardNumber}}}}
		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Internal server error")
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Payment details updated successfully"})
	}
}
func DeletePaymentDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("id")

		if userID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "invalid search index"})
			c.Abort()
			return
		}

		userObjectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Internal Server error")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.D{{Key: "_id", Value: userObjectID}}
		update := bson.D{{Key: "$unset", Value: bson.D{{Key: "payment_details", Value: ""}}}} // Удалить поле payment_details

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, "Wrong Command")
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Payment details deleted successfully"})
	}
}
