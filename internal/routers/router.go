package routers

import (
	"API.GOLANG.PROJECT_MEMORYBOX/internal/controller"
	"github.com/gofiber/fiber/v2"
)

func SetupRouter(app *fiber.App) {
	// Auth routes

	api := app.Group("/api")
	auth := api.Group("/auth")
	event := api.Group("/event")
	join := api.Group("/join")
	media := api.Group("/media")

	auth.Get("/GetUserByID/:uid", controller.GetUserByID)
	auth.Post("/login", controller.Login)
	auth.Post("/Regsiter", controller.Regsiter)

	event.Get("GetAllEvent", controller.GetAllEvent)
	event.Post("/UploadImageCover/:eid", controller.EventUploadImageCover)
	event.Post("CreateEvent", controller.EventCreate)
	event.Get("/GetEventsWithAttendees", controller.GetEventsWithAttendees)
	event.Get("/GetEventDetailWithAttendees/:eid", controller.GetEventDetailWithAttendees)
	event.Get("/EventGetMediaByID/:eid", controller.EventGetMediaByID)

	join.Post("/JoinEvent", controller.JoinEvent)
	media.Post("/upload/:eid/:uid", controller.UploadMediaAPI)
	// upload.Get("/images", controller.GetAllImagesAPI)
	// upload.Get("/image/:id", controller.GetImageByIDAPI)
	// upload.Delete("/image/:id", controller.DeleteImageAPI)
}
