package request

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	GoogleID string `json:"google_id"`
}

type GoogleAuthRequest struct {
	Email    string `json:"email"`
	GoogleID string `json:"google_id"`
}

type SendOTP struct {
	Email string `json:"email"`
}

type ChangePass struct {
	Email   string `json:"email"`
	Newpass string `json:"newpass"`
}

type OTPVerify struct {
	OTP   string `json:"otp"`
	Email string `json:"email"`
}
