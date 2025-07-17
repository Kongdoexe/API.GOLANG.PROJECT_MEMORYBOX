package main

import (
	"log"
	"os"

	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/routers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

const (
	uploadDir = "./uploads"
	maxSize   = 10 * 1024 * 1024
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
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	// config := controller.Config{
	// 	BucketName:      os.Getenv("BUCKETNAME"),
	// 	ProjectID:       os.Getenv("PROJECTID"),
	// 	CredentialsFile: os.Getenv("CREDENTIALSFILE"),
	// }

	// fileupload, err := controller.NewFileUpload(config)
	// if err != nil {
	// 	log.Fatalf("Failed to initialize file upload: %v", err)
	// }

	sqlDb, err := database.DB.DB()

	if err != nil {
		panic("Error in sql connection.")
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal("Cannot create upload directory:", err)
	}

	app.Static("/uploads", "./uploads")
	app.Static("/public", "./public")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"msg": "Hello, World!!!"})
	})

	defer sqlDb.Close()
	routers.SetupRouter(app)

	app.Listen(":5000")
}
