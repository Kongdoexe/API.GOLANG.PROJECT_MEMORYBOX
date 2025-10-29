package controller

import (
	"fmt"

	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/services"
	"github.com/gofiber/fiber/v2"
)

func GetUserByID(c *fiber.Ctx) error {
	uid := c.Params("uid")
	if uid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ไม่สามารถดำเนินการได้",
		})
	}
	response, err := services.GetUserByID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "ดึงข้อมูลสำเร็จ",
		"data":    response,
		"success": true,
	})
}

func GetUserByEmailAndGoogleID(c *fiber.Ctx) error {
	var email request.GoogleAuthRequest

	if err := c.BodyParser(&email); err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "ไม่สามารถดำเนินการได้",
		})
	}

	fmt.Print("Email: ", email.Email, " GoogleID: ", email.GoogleID, "\n")

	if email.Email == "" || email.GoogleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ไม่สามารถดำเนินการได้",
		})
	}

	_, err := services.GetUserByEmailAGoogleID(email.Email, email.GoogleID)

	if err != nil {
		return c.JSON(fiber.Map{
			"success": false,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

func RegisterGoolge(c *fiber.Ctx) error {
	var req models.User
	if err := c.BodyParser(&req); err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "ไม่สามารถดำเนินการได้",
		})
	}

	user, err := services.RegisterGoolge(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": user})
}

func Login(c *fiber.Ctx) error {
	var req request.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "ไม่สามารถดำเนินการได้",
		})
	}

	response, err, success := services.Login(&req)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "เข้าสู่ระบบสำเร็จ",
		"data":    response,
		"success": success,
	})
}

func Register(c *fiber.Ctx) error {
	var req models.User

	if err := c.BodyParser(&req); err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "ไม่สามารถดำเนินการได้",
		})
	}

	response, err, success := services.Register(&req)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "สมัครสมาชิกสำเร็จ",
		"data":    response,
		"success": success,
	})
}

func SendOTPEmail(c *fiber.Ctx) error {
	var req request.SendOTP
	if err := c.BodyParser(&req); err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "ไม่สามารถดำเนินการได้",
		})
	}

	err := services.SendOTPEmail(req)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "ส่งเสร็จสิ้น",
	})
}

func CheckOTP(c *fiber.Ctx) error {
	var req request.OTPVerify
	if err := c.BodyParser(&req); err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "ไม่สามารถดำเนินการได้",
		})
	}

	err := services.CheckOTP(req)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
			"success": false,
		})
	}

	return c.JSON(fiber.Map{
		"message": "สำเร็จ",
		"success": true,
	})
}

func ChangePass(c *fiber.Ctx) error {
	var req request.ChangePass
	if err := c.BodyParser(&req); err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "ไม่สามารถดำเนินการได้",
		})
	}

	err := services.ChangePass(req, false)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
			"success": false,
		})
	}

	return c.JSON(fiber.Map{
		"message": "เปลี่ยนรหัสสำเร็จ",
		"success": true,
	})
}

func ChangePassOTP(c *fiber.Ctx) error {
	var req request.ChangePass
	if err := c.BodyParser(&req); err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "ไม่สามารถดำเนินการได้",
		})
	}

	err := services.ChangePass(req, true)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
			"success": false,
		})
	}

	return c.JSON(fiber.Map{
		"message": "เปลี่ยนรหัสสำเร็จ",
		"success": true,
	})
}

func ChangeProfile(c *fiber.Ctx) error {
	var req request.ChangeProfile
	if err := c.BodyParser(&req); err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "ไม่สามารถดำเนินการได้",
		})
	}

	user, err := services.ChangeProfile(req)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
			"success": false,
		})
	}

	return c.JSON(fiber.Map{
		"message": "เปลี่ยนโปรไฟล์สำเร็จ",
		"data":    user,
	})
}

func UserUploadImageCover(c *fiber.Ctx) error {
	uid := c.Params("uid")

	if uid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ขาด UserID",
		})
	}

	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ไม่พบไฟล์ที่อัปโหลด",
		})
	}

	imageurl, err := services.UserUploadImageCover(file, uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":   "อัปโหลดสำเร็จ",
		"image_url": imageurl,
	})
}

func GetAllUserInSystem(c *fiber.Ctx) error {
	response, err := services.GetAllUserInSystem()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "ดึงข้อมูลสำเร็จ",
		"data":    response,
	})
}
