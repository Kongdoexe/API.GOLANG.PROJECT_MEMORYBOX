package repositories

import (
	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
)

func UserCreate(user *models.User) error {
	return database.DB.Create(user).Error
}

func UserGetAll() (*[]models.User, error) {
	var users []models.User

	if err := database.DB.Find(&users).Error; err != nil {
		return nil, err
	}

	return &users, nil
}

func UserFindByID(id string) (*models.User, error) {
	var user models.User

	if err := database.DB.Where("user_id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func UserFindByname(username string) *models.User {
	var user models.User

	if err := database.DB.Where("name LIKE ?", "%"+username+"%").First(&user).Error; err != nil {
		return nil
	}

	return &user
}

func UserFindByEmailAGoogleID(email, googleId string) (*models.User, error) {
	var user models.User

	if err := database.DB.Where("email = ? AND google_id = ?", email, googleId).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func UserFindByEmail(email string) (*models.User, error) {
	var user models.User

	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func UserFindPhoneNumber(phone string) (*models.User, error) {
	var user models.User

	if phone == "" {
		return &user, nil
	}

	if err := database.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func UserFindlastUID() (*models.User, error) {
	var user models.User

	if err := database.DB.Last(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func UserFindByGoogleID(googleId string) (*models.User, error) {
	var user models.User

	if err := database.DB.Where("google_id = ?", googleId).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func GetOTPByUserId(userId uint) *models.ResetPassToken {
	var resetpass models.ResetPassToken

	if err := database.DB.Where("user_id = ?", userId).First(&resetpass).Error; err != nil {
		return nil
	}

	return &resetpass
}

func CreateOTPByEmail(resetpass *models.ResetPassToken) error {
	return database.DB.Create(resetpass).Error
}

func UpdateOTPByEmail(resetpass *models.ResetPassToken) error {
	return database.DB.Save(resetpass).Error
}

func ChangeProfileUser(user *models.User) error {
	return database.DB.Save(user).Error
}

func UserUploadImageCover(uid int, imagePath string) (*models.User, error) {
	var user models.User

	// ค้นหา Event ตาม uid
	if err := database.DB.First(&user, uid).Error; err != nil {
		return nil, err
	}

	user.UserImage = imagePath

	if err := database.DB.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
