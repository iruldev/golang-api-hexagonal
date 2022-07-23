package service

import (
	"context"
	"github.com/iruldev/golang-api-hexagonal/model/web"
)

type CategoryService interface {
	Create(ctx context.Context, request web.CategoryCreateRequest) web.Categoryresponse
	Update(ctx context.Context, request web.CategoryUpdateRequest) web.Categoryresponse
	Delete(ctx context.Context, categoryId int)
	FindById(ctx context.Context, categoryId int) web.Categoryresponse
	FindAll(ctx context.Context) []web.Categoryresponse
}
