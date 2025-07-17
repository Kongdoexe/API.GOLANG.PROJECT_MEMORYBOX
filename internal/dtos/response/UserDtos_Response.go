package response

import "API.GOLANG.PROJECT_MEMORYBOX/internal/models"

type LoginResponse struct {
	User models.User `json:"user"`
}
