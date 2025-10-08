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
	noti := api.Group("/noti")
	//chat := event.Group("/chat")

	auth.Get("/GetUserByID/:uid", controller.GetUserByID)
	auth.Post("/login", controller.Login)
	auth.Post("/Register", controller.Register)
	auth.Post("/GetUserByEmailAndGoogleID", controller.GetUserByEmailAndGoogleID)

	auth.Post("/SendOTPEmail", controller.SendOTPEmail)
	auth.Post("/CheckOTP", controller.CheckOTP)
	auth.Post("/ChangePass", controller.ChangePass)

	event.Get("GetAllEvent", controller.GetAllEvent)
	event.Post("/UploadImageCover/:eid", controller.EventUploadImageCover)
	event.Post("CreateEvent", controller.EventCreate)
	event.Get("/GetEventsWithAttendees/:uid", controller.GetEventsWithAttendees)
	event.Get("/GetEventDetailWithAttendees/:eid", controller.GetEventDetailWithAttendees)
	event.Get("/EventGetListJoinUser/:eid/:uid", controller.EventGetListJoinUser)
	event.Get("/EventGetMediaByID/:eid", controller.EventGetMediaByID)
	event.Get("/EventCheckJoinUser/:eid/:uid", controller.EventCheckJoinUser)
	event.Get("/GetEventMainProfile/:uid", controller.GetEventMainProfile)

	// event.Get("/GetEventFavorite/:uid", controller.EventGetFavorite)
	event.Post("/EventFavorite", controller.EventFavorite)

	join.Post("/JoinEvent", controller.JoinEvent)
	media.Post("/upload/:eid/:uid", controller.UploadMediaAPI)

	noti.Post("/InsertTok", controller.InsertTokenNotification)
	noti.Post("/SendNoti", controller.SendNotification)
}
