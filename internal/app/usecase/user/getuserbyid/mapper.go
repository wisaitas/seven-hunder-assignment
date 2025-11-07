package getuserbyid

import "github.com/7-solutions/backend-challenge/internal/app/domain/entity"

func (s *service) mapEntityToResponse(user *entity.User) Response {
	return Response{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}
