package helper

import (
	"github.com/iruldev/golang-api-hexagonal/model/domain"
	"github.com/iruldev/golang-api-hexagonal/model/web"
)

func ToCategoryResponse(category domain.Category) web.Categoryresponse {
	return web.Categoryresponse{
		Id:   category.Id,
		Name: category.Name,
	}
}

func ToCategoryResponses(categories []domain.Category) []web.Categoryresponse {
	var Categoryresponses []web.Categoryresponse
	for _, category := range categories {
		Categoryresponses = append(Categoryresponses, ToCategoryResponse(category))
	}
	return Categoryresponses
}
