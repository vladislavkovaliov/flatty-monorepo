package main

import (
	"github.com/gin-gonic/gin"

	"github.com/jackc/pgx/v5/pgxpool"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"flatty-budget/go-api/http/handlers"
	"flatty-budget/go-api/internal/auth"
	"flatty-budget/go-api/internal/config"
	categoryrepo "flatty-budget/go-api/repos/category"
	expensestatsrepo "flatty-budget/go-api/repos/expense_stats"
	residentlocationrepo "flatty-budget/go-api/repos/resident_location"
	userrepo "flatty-budget/go-api/repos/user"
	categoryservice "flatty-budget/go-api/services/category"
	expensestatsservice "flatty-budget/go-api/services/expense_stats"
	expensesservice "flatty-budget/go-api/services/expenses"
	residentlocationservice "flatty-budget/go-api/services/resident_location"
	userservice "flatty-budget/go-api/services/user"

	_ "flatty-budget/go-api/docs"
)

func setupRouter(pool *pgxpool.Pool, expenseSvc *expensesservice.Service, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	authMw := auth.AuthMiddleware(pool)

	api := r.Group("/api")
	wireConfig(api)
	wireResidentLocation(api, pool, authMw)
	wireCategory(api, pool, authMw)
	wireExpenses(api, pool, expenseSvc, authMw)
	wireExpenseStats(api, pool, authMw)
	wireUser(api, pool, authMw)

	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}

func wireConfig(rg *gin.RouterGroup) {
	h := handlers.NewConfigHandler()

	rg.GET("/health", h.Health)
}

func wireResidentLocation(rg *gin.RouterGroup, pool *pgxpool.Pool, authMw gin.HandlerFunc) {

	repo := residentlocationrepo.NewPgxRepository(pool)
	svc := residentlocationservice.New(repo)

	h := handlers.NewResidentLocationHandler(svc)

	protected := rg.Group("", authMw)
	protected.GET("/resident-location", h.List)
	protected.POST("/resident-location", h.Create)
	protected.PUT("/resident-location/:id", h.Update)
	protected.DELETE("/resident-location/:id", h.Delete)
	protected.GET("/resident-location/count", h.Count)
}

func wireCategory(rg *gin.RouterGroup, pool *pgxpool.Pool, authMw gin.HandlerFunc) {

	repo := categoryrepo.NewPgxRepository(pool)
	svc := categoryservice.New(repo)

	h := handlers.NewCategoryHandler(svc)

	protected := rg.Group("", authMw)
	protected.GET("/categories", h.List)
	protected.POST("/categories", h.Create)
	protected.PUT("/categories/:id", h.Update)
	protected.DELETE("/categories/:id", h.Delete)
	protected.GET("/categories/count", h.Count)
}

func wireExpenses(rg *gin.RouterGroup, pool *pgxpool.Pool, svc *expensesservice.Service, authMw gin.HandlerFunc) {
	h := handlers.NewExpenseHandler(svc)

	protected := rg.Group("", authMw)
	protected.GET("/expenses", h.List)
	protected.POST("/expenses", h.Create)
	protected.PUT("/expenses/:id", h.Update)
	protected.DELETE("/expenses/:id", h.Delete)
	protected.GET("/expenses/count", h.Count)
}

func wireExpenseStats(rg *gin.RouterGroup, pool *pgxpool.Pool, authMw gin.HandlerFunc) {
	totalRepo := expensestatsrepo.NewPgxMonthlyTotalRepository(pool)
	avgRepo := expensestatsrepo.NewPgxMonthlyAverageRepository(pool)

	totalSvc := expensestatsservice.NewMonthlyTotalService(totalRepo)
	avgSvc := expensestatsservice.NewMonthlyAverageService(avgRepo)

	h := handlers.NewExpenseStatsHandler(totalSvc, avgSvc)

	protected := rg.Group("", authMw)
	protected.GET("/expenses/stats/totals", h.ListTotals)
	protected.GET("/expenses/stats/averages", h.ListAverages)
}

func wireUser(rg *gin.RouterGroup, pool *pgxpool.Pool, authMw gin.HandlerFunc) {
	userRepo := userrepo.NewPgxRepository(pool)
	userSvc := userservice.NewUserService(userRepo)

	h := handlers.NewUseHandler(userSvc)

	protected := rg.Group("", authMw)

	protected.GET("/user", h.List)
	protected.GET("/user/:id", h.GetUserByID)
}
