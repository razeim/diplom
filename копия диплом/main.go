package main

import (
	"beep/controllers"
	"beep/database"
	"beep/music"
	"beep/routes"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/gin-gonic/gin"
)

const vkCloudHotboxEndpoint = "https://hb.ru-msk.vkcs.cloud"
const defaultRegion = "us-east-1"

func main() {
	sess, _ := session.NewSession()
	svc := s3.New(sess, aws.NewConfig().WithEndpoint(vkCloudHotboxEndpoint).WithRegion(defaultRegion))
	bucket := "razeim"
	result, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		log.Fatalf("Unable to list items in bucket %q, %v", bucket, err)
	} else {
		// итерирование по объектам
		for _, item := range result.Contents {
			log.Printf("Object: %s, size: %d\n", aws.StringValue(item.Key), aws.Int64Value(item.Size))
		}
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"

	}
	app := controllers.NewApplication(database.ProductData(database.Client, "Products"), database.UserData(database.Client, "Users"))

	router := gin.New()
	router.Use(gin.Logger())
	routes.UserRoutes(router)
	//router.Use(middleware.Authentication())
	router.LoadHTMLFiles("./smth/audio.html", "./smth/upload.html", "./smth/success.html", "./smth/login.html")
	router.Static("/static", "./smth")
	router.GET("/upload", func(c *gin.Context) {
		userId := c.Query("id")
		c.HTML(http.StatusOK, "upload.html", gin.H{"userId": userId})
	})
	router.GET("/users/signup", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	router.POST("/upload", music.UploadedFiles())
	router.GET("/addtocart", app.AddToCart())
	router.GET("/removeitem", app.RemoteItem())
	router.GET("/listcart", controllers.GetItemFromCart())
	router.POST("/addpaymentdetails", controllers.AddPaymentDetails())
	router.PUT("/editpaymentdetails", controllers.EditPaymentDetails())
	router.GET("/deletpaymentdetails", controllers.DeletePaymentDetails())
	router.GET("/cartcheckout", app.BuyFromCart())
	router.GET("/instantbuy", app.InstantBuy())
	router.GET("/audio", music.AudioPlay())
	log.Fatal(router.Run(":" + port))
}
