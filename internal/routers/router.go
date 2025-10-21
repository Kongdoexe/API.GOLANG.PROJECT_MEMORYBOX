package routers

import (
	"API.GOLANG.PROJECT_MEMORYBOX/internal/controller"
	"github.com/gofiber/fiber/v2"
)

func SetupRouter(app *fiber.App) {

	api := app.Group("/api")
	auth := api.Group("/auth")
	event := api.Group("/event")
	join := api.Group("/join")
	media := api.Group("/media")
	noti := api.Group("/noti")
	chat := api.Group("/chat")
	//chat := event.Group("/chat")

	auth.Get("/GetAllUserInSystem", controller.GetAllUserInSystem)
	auth.Get("/GetUserByID/:uid", controller.GetUserByID)
	auth.Post("/login", controller.Login)
	auth.Post("/Register", controller.Register)
	auth.Post("/GetUserByEmailAndGoogleID", controller.GetUserByEmailAndGoogleID)
	auth.Post("/UserUploadImageCover/:uid", controller.UserUploadImageCover)

	auth.Post("/SendOTPEmail", controller.SendOTPEmail)
	auth.Post("/CheckOTP", controller.CheckOTP)
	auth.Post("/ChangePassOTP", controller.ChangePassOTP)
	auth.Post("/ChangePass", controller.ChangePass)

	auth.Post("/ChangeProfile", controller.ChangeProfile)

	event.Get("GetAllEvent", controller.GetAllEvent)
	event.Post("/UploadImageCover/:eid", controller.EventUploadImageCover)
	event.Post("CreateEvent", controller.EventCreate)
	event.Delete("/EventDelete/:eid", controller.EventDelete)
	event.Get("/GetEventsWithAttendees/:uid", controller.GetEventsWithAttendees)
	event.Get("/GetEventDetailWithAttendees/:eid", controller.GetEventDetailWithAttendees)
	event.Get("/EventGetListJoinUser/:eid/:uid", controller.EventGetListJoinUser)
	event.Get("/EventGetMediaByID/:eid", controller.EventGetMediaByID)
	event.Get("/EventCheckJoinUser/:eid/:uid", controller.EventCheckJoinUser)
	event.Get("/GetEventMainProfile/:uid", controller.GetEventMainProfile)
	event.Get("/EventGetFavorites/:uid", controller.EventGetFavorites)
	event.Get("/GetEventCalendar/:uid", controller.GetEventCalendar)
	event.Get("/GetEventJoin/:uid", controller.GetEventJoin)

	event.Post("/EventFavorite", controller.EventFavorite)

	join.Post("/JoinEvent", controller.JoinEvent)
	join.Put("/JoinBlocked", controller.JoinBlocked)
	join.Post("/CheckBlocked", controller.CheckBlocked)

	media.Post("/upload/:eid/:uid", controller.UploadMediaAPI)
	media.Delete("/DeleteMediaByID", controller.DeleteMediaByID)

	noti.Post("/InsertTok", controller.InsertTokenNotification)
	noti.Post("/event", controller.SendNotificationEvent)
	noti.Post("/chat", controller.SendNotificationChat)
	noti.Get("/GetUserNotification/:uid", controller.GetUserNotification)
	noti.Put("/UpdateIsReadNotification/:nid", controller.UpdateIsReadNotification)

	chat.Post("/InsertMessage", controller.InsertMessage)
	chat.Post("/GetMessage", controller.GetMessage)
}
