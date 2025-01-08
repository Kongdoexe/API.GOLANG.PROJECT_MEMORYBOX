package controller

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/models"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/rand"
	"gopkg.in/gomail.v2"
)

type EmailRequest struct {
	Email string `json:"email"`
}

type OTPResponse struct {
	UserID int `json:"user_id"`
	OTP    int `json:"otp"`
}

type NewPassword struct {
	UserID   int    `json:"user_id"`
	Password string `json:"password"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func SelectAllUser(c *fiber.Ctx) error {
	var users []models.User

	result := database.DBconn.
		Preload("Events").
		Preload("Joins").
		Preload("Favorites").
		Preload("Media").
		Find(&users)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch users data",
			"details": result.Error.Error(),
		})
	}

	for i := range users {
		users[i].Password = ""
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   users,
		"count":  len(users),
	})
}

func RegisterRest(c *fiber.Ctx) error {

	var gmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@gmail\.com$`)

	var member models.User
	if err := c.BodyParser(&member); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cannot parse JSON",
		})
	}

	if !gmailRegex.MatchString(member.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Email must be a Gmail address",
		})
	}

	if err := database.DBconn.Where("email = ?", member.Email).First(&member).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Email already exists",
		})
	}

	if _, err := strconv.Atoi(member.Phone); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Phone number must be number",
		})
	}

	if phone := len(member.Phone); phone != 10 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Phone number must be 10",
		})
	}

	if err := database.DBconn.Where("phone = ?", member.Phone).First(&member).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Phone already exists",
		})
	}

	hash, err := HashPassword(member.Password)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error hashing password"})
	}

	member.UserImage = "https://firebasestorage.googleapis.com/v0/b/project-memorybox.firebasestorage.app/o/profile%2Fprofile.png?alt=media&token=33bff74a-4942-442a-a8ad-0f76c448bdbe"
	member.Password = hash
	member.IsNotification = true
	member.GoogleID = nil

	if err := database.DBconn.Create(&member).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cannot create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"user":    member,
	})
}

func RegisterGoogle(c *fiber.Ctx) error {
	var member models.User
	if err := c.BodyParser(&member); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cannot parse JSON",
		})
	}

	if err := database.DBconn.Where("email = ?", member.Email).First(&member).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Email already exists",
		})
	}

	if _, err := strconv.Atoi(member.Phone); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Phone number must be number",
		})
	}

	if phone := len(member.Phone); phone != 10 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Phone number must be 10",
		})
	}

	if err := database.DBconn.Where("phone = ?", member.Phone).First(&member).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Phone already exists",
		})
	}

	hash, err := HashPassword(member.Password)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error hashing password"})
	}

	member.Password = hash
	member.IsNotification = true

	if err := database.DBconn.Create(&member).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cannot create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"user":    member,
	})
}

func LoginRest(c *fiber.Ctx) error {
	var member struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&member); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cannot parse JSON",
		})
	}

	var user models.User
	if err := database.DBconn.Where("email = ?", member.Email).First(&user).Preload("Events").
		Preload("Joins").
		Preload("Favorites").
		Preload("Media").Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	if !CheckPasswordHash(member.Password, user.Password) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Incorrect password",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"data":    user,
	})
}

func LoginGoogle(c *fiber.Ctx) error {
	var member struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&member); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cannot parse JSON",
		})
	}

	var user models.User
	if err := database.DBconn.Where("email = ? and google_id is not null", member.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
	})
}

// **
func UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")

	var member models.User
	if err := c.BodyParser(&member); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cannot parse JSON",
		})
	}

	var user models.User
	if err := database.DBconn.Where("user_id = ?", id).First(&member).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	if err := database.DBconn.Where("email = ? AND user_id != ?", member.Email, id).First(&user).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Email already exists",
		})
	}

	if _, err := strconv.Atoi(member.Phone); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Phone number must be number",
		})
	}

	if phone := len(member.Phone); phone != 10 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Phone number must be 10",
		})
	}

	if err := database.DBconn.Where("phone = ? AND user_id != ?", member.Phone, id).First(&user).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Phone already exists",
		})
	}

	if err := database.DBconn.Save(&member).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cannot update user",
			"error":   member,
		})
	}

	return c.JSON(fiber.Map{
		"message": "User updated successfully",
		"user":    member,
	})
}

// **
func DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var member models.User

	result := database.DBconn.Delete(&member, "user_id = ?", id)

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}

func generateOTP() string {
	rand.Seed(uint64(time.Now().UnixNano()))
	// Generate 6-digit OTP
	otp := rand.Intn(900000) + 100000
	return fmt.Sprintf("%d", otp)
}

func SendEmailOTP(c *fiber.Ctx) error {
	var req EmailRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var member models.User
	if err := database.DBconn.Where("email = ?", req.Email).First(&member).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	var tokenPrePass models.ResetPassToken
	if err := database.DBconn.Where("user_id = ?", member.ID).Last(&tokenPrePass).Error; err == nil {
		if time.Now().Before(tokenPrePass.Expire) {
			return c.JSON(fiber.Map{
				"message": "OTP is still valid",
			})
		}
	}

	otp := generateOTP()

	tokenRePass := models.ResetPassToken{
		UserID:    member.ID,
		Token:     otp,
		Expire:    time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := database.DBconn.Create(&tokenRePass).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save OTP token",
		})
	}

	m := gomail.NewMessage()
	m.SetHeader("From", "memoryboxapplication@gmail.com")
	m.SetHeader("To", req.Email)
	m.SetHeader("Subject", "Your Verification Code")
	m.SetHeader("List-Unsubscribe", "<mailto:unsubscribe@memoryboxapplication.com>")

	m.AddAlternative("text/html", fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
	</head>
	<body>
		<h2>MemoryBox Application - OTP Verification</h2>
		<p>Thank you for using MemoryBox!</p>
		<p>Your OTP code is: <strong style="color: red">%s</strong></p>
		<p>Please enter this code within 5 minutes to verify your account.</p>
		<p><strong>If you didn't request this code, please ignore this email.</strong></p>
	</body>
	</html>
	`, otp))

	d := gomail.NewDialer(
		"smtp.gmail.com",
		587,
		"memoryboxapplication@gmail.com",
		"ailkpqnwoqsvuege",
	)

	if err := d.DialAndSend(m); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send email",
		})
	}

	return c.JSON(fiber.Map{
		"message": "OTP sent successfully",
	})
}

func CheckOTPToken(c *fiber.Ctx) error {
	var req OTPResponse

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var token models.ResetPassToken
	if err := database.DBconn.Where("token = ? AND user_id = ?", req.OTP, req.UserID).Last(&token).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid OTP",
		})
	}

	if time.Now().After(token.Expire) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "OTP has expired",
		})
	}

	return c.JSON(fiber.Map{
		"message": "OTP Check Verification completed",
		"user_id": token.UserID,
	})
}

func ResetPassword(c *fiber.Ctx) error {
	var req NewPassword

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var member models.User
	if err := database.DBconn.Where("user_id = ?", req.UserID).First(&member).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User not Found",
		})
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error hashing password",
		})
	}

	member.Password = hash

	if err := database.DBconn.Save(&member).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Cannot update password",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Reset Password successfully",
	})
}
