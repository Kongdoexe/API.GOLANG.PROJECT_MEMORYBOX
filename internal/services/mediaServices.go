package services

import (
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/response"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/repositories"
)

func GetMediaInfo(mediaId uint) (*response.MediaResponse, error) {
	media, user, err := repositories.GetMediaByID(mediaId)
	if err != nil {
		return nil, err
	}

	response := &response.MediaResponse{
		MediaID:    mediaId,
		ImageURL:   media.FileURL,
		UploadedAt: media.UploadTime,
		TakenAt:    media.DetailTime,
	}

	response.UploadedBy.UserID = user.ID
	response.UploadedBy.UserName = user.Name

	return response, nil
}
