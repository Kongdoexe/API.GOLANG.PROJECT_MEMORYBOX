package main

import (
	"log"
	"os"

	"API.GOLANG.PROJECT_MEMORYBOX/controller"
	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/routers"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func init() {

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error in loading .env file.")
	}

	database.InitDB()
}

func main() {
	app := fiber.New(fiber.Config{
		BodyLimit: 4048 * 1024 * 1024,
	})

	config := controller.Config{
		BucketName:      os.Getenv("BUCKETNAME"),
		ProjectID:       os.Getenv("PROJECTID"),
		CredentialsFile: os.Getenv("CREDENTIALSFILE"),
	}

	fileupload, err := controller.NewFileUpload(config)
	if err != nil {
		log.Fatalf("Failed to initialize file upload: %v", err)
	}

	sqlDb, err := database.DBconn.DB()

	if err != nil {
		panic("Error in sql connection.")
	}

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"msg": "Hello, World!!!"})
	})

	defer sqlDb.Close()
	routers.SetupRouter(app, fileupload)

	app.Listen(":3000")
}
