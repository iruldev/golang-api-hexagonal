package main

import (
	"github.com/iruldev/golang-api-hexagonal/app"
	"github.com/iruldev/golang-api-hexagonal/controller"
	"github.com/iruldev/golang-api-hexagonal/helper"
	"github.com/iruldev/golang-api-hexagonal/middleware"
	"github.com/iruldev/golang-api-hexagonal/repository"
	"github.com/iruldev/golang-api-hexagonal/service"
	"net/http"

	"github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db := app.NewDB()
	validate := validator.New()
	categoryRepository := repository.NewCategoryRepository()
	categoryService := service.NewCategoryService(categoryRepository, db, validate)
	categoryController := controller.NewCategoryController(categoryService)

	router := app.NewRouter(categoryController)

	server := http.Server{
		Addr:    "localhost:3000",
		Handler: middleware.NewAuthMiddleware(router),
	}

	err := server.ListenAndServe()
	helper.PanicIfError(err)
}
