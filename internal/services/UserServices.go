package services

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

func generateOTP() string {
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(900000) + 100000
	return fmt.Sprintf("%d", otp)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetLastID() (*models.User, error) {
	user, err := repositories.UserFindlastUID()
	if err != nil {
		return nil, errors.New("ไม่สามารถค้นหาได้")
	}

	return user, nil
}

func Login(req *request.LoginRequest) (*models.User, error, bool) {
	user, err := repositories.UserFindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("ไม่พบผู้ใช้"), false
	}

	if req.GoogleID != "" {
		_, err = repositories.UserFindByGoogleID(req.GoogleID)
		if err == nil {
			return user, nil, true
		}
	}

	if !CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("รหัสผ่านไม่ถูกต้อง"), false
	}

	return user, nil, true
}

func Regsiter(req *models.User) (*models.User, error, bool) {
	_, err := repositories.UserFindByEmail(req.Email)
	if err == nil {
		return nil, errors.New("มีผู้ใช้อีเมลนี้แล้ว"), false
	}

	_, err = repositories.UserFindPhoneNumber(req.Phone)
	if err == nil {
		return nil, errors.New("มีผู้ใช้เบอร์นี้แล้ว"), false
	}

	if req.GoogleID == nil {
		req.UserImage = "http://117.18.125.19/images/profile.png"
		req.IsNotification = true
		hashedPassword, err := HashPassword(req.Password)
		if err != nil {
			return nil, errors.New("ไม่สามารถเข้ารหัสรหัสผ่านได้"), false
		}
		req.Password = hashedPassword

		if err = repositories.UserCreate(req); err != nil {
			return nil, errors.New("ไม่สามารถสร้างผู้ใช้ได้"), false
		}
	} else if req.GoogleID != nil {
		req.IsNotification = true
		if err = repositories.UserCreate(req); err != nil {
			return nil, errors.New("ไม่สามารถสร้างผู้ใช้ได้"), false
		}
	}

	return req, nil, true
}

func GetUserByID(req string) (*models.User, error) {
	user, err := repositories.UserFindByID(req)
	if err != nil {
		return nil, errors.New("ไม่พบผู้ใช้")
	}

	return user, nil
}
