package getusers

import "github.com/7-solutions/backend-challenge/internal/app/domain/entity"

func (s *service) mapEntityToResponse(entities []*entity.User) []Response {
	responses := make([]Response, len(entities))
	for i, entity := range entities {
		responses[i] = Response{
			ID:    entity.ID,
			Name:  entity.Name,
			Email: entity.Email,
		}
	}

	return responses
}
