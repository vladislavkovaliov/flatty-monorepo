package main

import (
	"github.com/gin-gonic/gin"

	"github.com/jackc/pgx/v5/pgxpool"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"flatty-budget/go-api/http/handlers"
	categoryrepo "flatty-budget/go-api/repos/category"
	categoryservice "flatty-budget/go-api/services/category"
	expensesrepo "flatty-budget/go-api/repos/expenses"
	expensesservice "flatty-budget/go-api/services/expenses"
	residentlocationrepo "flatty-budget/go-api/repos/resident_location"
	residentlocationservice "flatty-budget/go-api/services/resident_location"

	_ "flatty-budget/go-api/docs"
)

func setupRouter(pool *pgxpool.Pool) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	wireConfig(api)
	wireResidentLocation(api, pool)
	wireCategory(api, pool)
	wireExpenses(api, pool)

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

	rg.GET("/resident-location", h.List)
	rg.POST("/resident-location", h.Create)
	rg.PUT("/resident-location/:id", h.Update)
	rg.DELETE("/resident-location/:id", h.Delete)
	rg.GET("/resident-location/count", h.Count)
}

func wireCategory(rg *gin.RouterGroup, pool *pgxpool.Pool) {

	repo := categoryrepo.NewPgxRepository(pool)
	svc := categoryservice.New(repo)

	h := handlers.NewCategoryHandler(svc)

	rg.GET("/categories", h.List)
	rg.POST("/categories", h.Create)
	rg.PUT("/categories/:id", h.Update)
	rg.DELETE("/categories/:id", h.Delete)
	rg.GET("/categories/count", h.Count)
}

func wireExpenses(rg *gin.RouterGroup, pool *pgxpool.Pool) {

	repo := expensesrepo.NewPgxRepository(pool)
	svc := expensesservice.New(repo)

	h := handlers.NewExpenseHandler(svc)

	rg.GET("/expenses", h.List)
	rg.POST("/expenses", h.Create)
	rg.PUT("/expenses/:id", h.Update)
	rg.DELETE("/expenses/:id", h.Delete)
	rg.GET("/expenses/count", h.Count)
}
