package services

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/repositories"
	"github.com/resend/resend-go/v2"
	"golang.org/x/crypto/bcrypt"
)

var gmailOnlyRegex = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9._%+\-]{0,62}[a-z0-9])?@gmail\.com$`)
var nameRegex = regexp.MustCompile(`^.{2,50}$`)
var thPhoneRegex = regexp.MustCompile(`^0\d{9}$`)

// Go's regexp does not support lookahead, so we check password strength in code.
func isPasswordStrong(password string) bool {
	if len(password) < 8 || len(password) > 64 {
		return false
	}
	if strings.Contains(password, " ") {
		return false
	}
	var hasLower, hasUpper, hasDigit, hasSpecial bool
	specialChars := "~!@#$%^&*()_-+={}[]|\\:;\"'<>,.?/"
	for _, c := range password {
		switch {
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= '0' && c <= '9':
			hasDigit = true
		case strings.ContainsRune(specialChars, c):
			hasSpecial = true
		}
	}
	return hasLower && hasUpper && hasDigit && hasSpecial
}

func sanitizeUserInput(u *models.User) {
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	u.Phone = strings.TrimSpace(u.Phone)
	u.Name = strings.TrimSpace(u.Name)
}

func generateOTP() string {
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(9000) + 1000
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

func Register(req *models.User) (*models.User, error, bool) {
	if req == nil {
		return nil, errors.New("คำขอว่างเปล่า"), false
	}

	if req.Password == "" || !isPasswordStrong(req.Password) {
		return nil, errors.New("รหัสผ่านต้องยาวอย่างน้อย 8 ตัว และมีตัวพิมพ์ใหญ่/เล็ก/ตัวเลข/อักขระพิเศษ อย่างละอย่างน้อย 1 และห้ามมีช่องว่าง"), false
	}
	if req.Email == "" {
		return nil, errors.New("กรุณากรอกอีเมล"), false
	}

	isGoogleSignup := req.GoogleID != nil && strings.TrimSpace(*req.GoogleID) != ""

	if _, err := repositories.UserFindByEmail(req.Email); err == nil {
		return nil, errors.New("มีผู้ใช้อีเมลนี้แล้ว"), false
	}

	if !isGoogleSignup {
		if !gmailOnlyRegex.MatchString(req.Email) {
			return nil, errors.New("อีเมลต้องเป็น gmail.com เท่านั้น และรูปแบบต้องถูกต้อง"), false
		}

		if req.Name == "" || !nameRegex.MatchString(req.Name) {
			return nil, errors.New("กรุณากรอกชื่อให้ถูกต้อง (2–50 ตัวอักษร)"), false
		}
		if req.Phone == "" || !thPhoneRegex.MatchString(req.Phone) {
			return nil, errors.New("กรุณากรอกเบอร์โทรให้ถูกต้อง (ตัวอย่าง 08xxxxxxxx)"), false
		}
		if req.Password == "" || !isPasswordStrong(req.Password) {
			return nil, errors.New("รหัสผ่านต้องยาวอย่างน้อย 8 ตัว และมีตัวพิมพ์ใหญ่/เล็ก/ตัวเลข/อักขระพิเศษ อย่างละอย่างน้อย 1 และห้ามมีช่องว่าง"), false
		}

		if _, err := repositories.UserFindPhoneNumber(req.Phone); err == nil {
			return nil, errors.New("มีผู้ใช้เบอร์นี้แล้ว"), false
		}

		if strings.TrimSpace(req.UserImage) == "" {
			req.UserImage = "http://117.18.125.19/images/profile.png"
		}
		req.IsNotification = 1

		hashedPassword, err := HashPassword(req.Password)
		if err != nil {
			return nil, errors.New("ไม่สามารถเข้ารหัสรหัสผ่านได้"), false
		}
		req.Password = hashedPassword
		if err := repositories.UserCreate(req); err != nil {
			return nil, errors.New("ไม่สามารถสร้างผู้ใช้ได้"), false
		}

	} else {
		req.IsNotification = 1
		req.Password = ""
		if err := repositories.UserCreate(req); err != nil {
			return nil, errors.New("ไม่สามารถสร้างผู้ใช้ได้"), false
		}
	}

	req.Password = ""
	return req, nil, true
}

func GetUserByID(req string) (*models.User, error) {
	user, err := repositories.UserFindByID(req)
	if err != nil {
		return nil, errors.New("ไม่พบผู้ใช้")
	}

	return user, nil
}

func GetUserByEmailAGoogleID(req string, goolge_id string) (*models.User, error) {
	user, err := repositories.UserFindByEmailAGoogleID(req, goolge_id)
	if err != nil {
		return nil, errors.New("ไม่พบผู้ใช้")
	}

	return user, nil
}

func SendOTPEmail(req request.SendOTP) error {
	apiResend := os.Getenv("apiKeyRend")
	client := resend.NewClient(apiResend)
	var reqResetPassModel *models.ResetPassToken

	user, err := repositories.UserFindByEmail(req.Email)
	if err != nil {
		return errors.New("ไม่พบผู้ใช้")
	}

	if user.GoogleID != nil {
		return errors.New("คุณล็อคอินด้วย Google ไม่สามารถรับ OTP ได้")
	}

	res := repositories.GetOTPByUserId(user.ID)

	if res != nil && time.Now().Before(res.Expire) {
		return errors.New("กรุณารออีก 1 นาที")
	}

	otp := generateOTP()

	reqResetPassModel = &models.ResetPassToken{
		UserID: user.ID,
		Token:  otp,
		Expire: time.Now().Add(1 * time.Minute),
	}

	if res == nil {
		err = repositories.CreateOTPByEmail(reqResetPassModel)
	} else {
		reqResetPassModel.ID = res.ID
		err = repositories.UpdateOTPByEmail(reqResetPassModel)
	}

	if err != nil {
		return errors.New("ไม่สามารถบันทึก OTP ลงฐานข้อมูลได้")
	}

	htmlBody := fmt.Sprintf(`
		<h2 style="color:#800080;">MemoryBox Application - OTP Verification</h2>
		<p>Thank you for using MemoryBox!</p>
		<p>Your OTP code is: <strong style="color:red;">%s</strong></p>
		<br/>
		<p>Please enter this code within 1 minutes to verify your account.</p>
		<p style="color:gray;">If you didn't request this code, please ignore this email.</p>
	`, otp)

	// ตั้งค่าข้อมูลอีเมล
	params := &resend.SendEmailRequest{
		From:    "MemoryBox <MemoryBox@k0n4n4p4.site>", // อีเมลที่ Verified แล้วใน Resend
		To:      []string{req.Email},                   // ผู้รับ
		Subject: "MemoryBox Application - OTP Verification",
		Html:    htmlBody,
	}

	// ส่งอีเมล
	_, err = client.Emails.Send(params)
	if err != nil {
		fmt.Println(err.Error())
		return errors.New("ไม่สามารถส่งอีเมลได้")
	}

	return nil
}

func CheckOTP(req request.OTPVerify) error {
	user, err := repositories.UserFindByEmail(req.Email)
	if err != nil {
		return errors.New("ไม่พบผู้ใช้")
	}

	res := repositories.GetOTPByUserId(user.ID)
	if res != nil && time.Now().After(res.Expire) {
		return errors.New("OTP หมดอายุกรุณาขอใหม่")
	}

	if res.Token != req.OTP {
		return errors.New("OTP ไม่ถูกต้อง")
	}

	return nil
}

func ChangePass(req request.ChangePass) error {
	user, err := repositories.UserFindByEmail(req.Email)
	if err != nil {
		return errors.New("ไม่พบผู้ใช้")
	}

	hashedPassword, err := HashPassword(req.Newpass)
	if err != nil {
		return errors.New("ไม่สามารถเข้ารหัสรหัสผ่านได้")
	}
	user.Password = hashedPassword

	err = repositories.ChangePass(user)
	if err != nil {
		return errors.New("ไม่สามารถเปลี่ยนรหัสผ่านได้")
	}

	return nil
}
