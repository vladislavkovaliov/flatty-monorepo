package main

import (
	"github.com/gin-gonic/gin"

	"github.com/jackc/pgx/v5/pgxpool"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"flatty-budget/go-api/http/handlers"
	residentlocationrepo "flatty-budget/go-api/repos/resident_location"
	residentlocationservice "flatty-budget/go-api/services/resident_location"

	_ "flatty-budget/go-api/docs"
)

func setupRouter(pool *pgxpool.Pool) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	wireConfig(api)
	wireResidentLocation(api, pool)

	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}

func wireConfig(rg *gin.RouterGroup) {
	h := handlers.NewConfigHandler()

	rg.GET("/health", h.Health)
}

func wireResidentLocation(rg *gin.RouterGroup, pool *pgxpool.Pool) {

	repo := residentlocationrepo.NewPgxRepository(pool)
	svc := residentlocationservice.New(repo)

	h := handlers.NewResidentLocationHandler(svc)

	rg.GET("/resident-location/count", h.Count)
}
