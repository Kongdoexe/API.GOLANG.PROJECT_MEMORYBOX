package routers

import (
	"API.GOLANG.PROJECT_MEMORYBOX/controller"
	"github.com/gofiber/fiber/v2"
)

func SetupRouter(app *fiber.App, fileUpload *controller.FileUpload) {
	// Auth routes

	api := app.Group("/api")
	auth := api.Group("/auth")
	event := api.Group("/event")
	upload := api.Group("/upload")

	app.Get("/users", controller.SelectAllUser)

	auth.Post("/registerRest", controller.RegisterRest)
	auth.Post("/registerGoogle", controller.RegisterGoogle)
	auth.Post("/loginRest", controller.LoginRest)
	auth.Post("/loginGoogle", controller.LoginGoogle)
	auth.Delete("/delete/:id", controller.DeleteUser)
	auth.Put("/update/:id", controller.UpdateUser)
	auth.Post("/sendemailotp", controller.SendEmailOTP)
	auth.Post("/check-otp-token", controller.CheckOTPToken)
	auth.Post("/reset-password", controller.ResetPassword)

	event.Get("/events", controller.SelectAllEvent)
	event.Post("IncreaseEvent", controller.IncreaseEvent)
	event.Post("JoinEvent", controller.JoinEvent)

	upload.Post("/uploadMediaEvent/:eid", fileUpload.HandleMultipleUpload)
	upload.Post("/uploadCoverImageEvent/:eid", fileUpload.HandleSingleUpload)
	upload.Post("/uploadCoverImageUser/:uid", fileUpload.HandleSingleUploadCoverImageUser)
	upload.Delete("/delete", fileUpload.DeleteFile)
	// upload.Post("/generate-qr-code/:eid", fileUpload.GenerateQR) ยังไม่ได้ทำ ทำยังไม่เสร็จ
}
