package music

import (
	"beep/database"
	"beep/models"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	vkCloudHotboxEndpoint = "https://hb.vkcs.cloud"
	defaultRegion         = "ru-msk"
)

var UserCollection *mongo.Collection = database.UserData(database.Client, "Users")
var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")

func AudioPlay() gin.HandlerFunc {
	return func(c *gin.Context) {
		audioURL := "https://razeim.hb.ru-msk.vkcs.cloud/РУСИК.mp3"
		c.HTML(http.StatusOK, "audio.html", gin.H{"audioURL": audioURL})

	}
}

func UploadedFiles() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		err := c.Request.ParseMultipartForm(10 << 20) // разбор формы с максимальным размером 10 MB
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"message": err.Error()})
			return
		}
		imageFile, imageHeader, err := c.Request.FormFile("image")
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"message": err.Error()})
			return
		}
		defer imageFile.Close()

		audioFile, audioHeader, err := c.Request.FormFile("audio")
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"message": err.Error()})
			return
		}
		defer audioFile.Close()

		imageURL := fmt.Sprintf("/temp/%s", imageHeader.Filename)
		audioURL := fmt.Sprintf("/temp/%s", audioHeader.Filename)
		err = uploadFileToVKStorage(c, imageURL, imageFile)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"message": err.Error()})
			return
		}

		err = uploadFileToVKStorage(c, audioURL, audioFile)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"message": err.Error()})
			return
		}

		userID := c.PostForm("id")
		if userID == "" {
			c.HTML(http.StatusBadRequest, "error.html", "user id is empty")
			fmt.Println("нет id")
			return
		}
		var newProduct models.Product
		newProduct.Product_ID = primitive.NewObjectID()
		newProduct.Author_ID, err = primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.IndentedJSON(500, "Internal server error")

		}
		priceStr := c.PostForm("price")
		price, err := strconv.ParseUint(priceStr, 10, 64)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"message": "Invalid price format"})
			return
		}
		newProduct.Price = aws.Uint64(price)
		newProduct.Product_Name = &audioHeader.Filename
		fileLink := fmt.Sprintf("https://razeim.hb.ru-msk.vkcs.cloud/%s", audioHeader.Filename)
		imageLink := fmt.Sprintf("https://razeim.hb.ru-msk.vkcs.cloud/%s", imageHeader.Filename)
		newProduct.File_Link = &fileLink
		newProduct.Image = &imageLink
		_, err = ProductCollection.InsertOne(ctx, newProduct)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert product into database"})
			return
		}

		userObjectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(500, "Internal Server error")
			return
		}
		var user models.User
		err = UserCollection.FindOne(ctx, bson.M{"_id": userObjectID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Failed to find user")
			return
		}
		user.Uploaded_Files = append(user.Uploaded_Files, newProduct.Product_ID)
		update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "uploaded_files", Value: user.Uploaded_Files}}}}
		_, err = UserCollection.UpdateOne(ctx, bson.M{"_id": userObjectID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Failed to update user")
			return
		}

		c.HTML(http.StatusOK, "success.html", nil)
	}
}

func uploadFileToVKStorage(c *gin.Context, fileURL string, uploadedFile multipart.File) error {
	sess, err := session.NewSession(&aws.Config{
		Endpoint: aws.String(vkCloudHotboxEndpoint),
		Region:   aws.String(defaultRegion),
	})
	if err != nil {
		return err
	}

	svc := s3.New(sess)
	bucket := "razeim"

	// Определение имени файла на основе URL
	fileName := filepath.Base(fileURL)

	// Загрузка файла в бакет
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
		Body:   uploadedFile,              // Используем загруженный файл
		ACL:    aws.String("public-read"), // Установка ACL на public-read
	})
	if err != nil {
		return err
	}

	log.Printf("File %s uploaded successfully to bucket %q", fileName, bucket)
	return nil
}
func AddProduct(c *gin.Context, userID string) {

}
