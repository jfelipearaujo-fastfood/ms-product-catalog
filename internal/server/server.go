package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jfelipearaujo-org/ms-product-catalog/internal/adapter/database"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/environment"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/handler/category/create_category_handler"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/handler/category/delete_category_handler"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/handler/category/get_categories_handler"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/handler/category/get_category_handler"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/handler/health_handler"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/handler/product/create_product_handler"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/handler/product/delete_product_handler"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/handler/product/get_product_handler"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/handler/product/get_products_handler"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/provider/time_provider"
	category_repository "github.com/jfelipearaujo-org/ms-product-catalog/internal/repository/category"
	product_repository "github.com/jfelipearaujo-org/ms-product-catalog/internal/repository/product"
	token "github.com/jfelipearaujo-org/ms-product-catalog/internal/server/middlewares"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/service/category/create_category"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/service/category/delete_category"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/service/category/get_categories"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/service/category/get_category"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/service/product/create_product"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/service/product/delete_product"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/service/product/get_product"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/service/product/get_products"
	"github.com/jfelipearaujo-org/ms-product-catalog/internal/shared/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	Config *environment.Config
	db     database.DatabaseService
}

func NewServer(config *environment.Config) *http.Server {
	server := &Server{
		Config: config,
		db:     database.NewDatabase(config),
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", server.Config.ApiConfig.Port),
		Handler:      server.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}

func (server *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(logger.Middleware())
	e.Use(middleware.Recover())

	server.registerHealthCheck(e)

	group := e.Group(fmt.Sprintf("/api/%s", server.Config.ApiConfig.ApiVersion))

	group.Use(token.Middleware())
	server.registerCategoryRoutes(group)
	server.registerProductRoutes(group)

	return e
}

func (server *Server) registerHealthCheck(e *echo.Echo) {
	healthHandler := health_handler.NewHandler(server.db)

	e.GET("/health", healthHandler.Handle)
}

func (server *Server) registerCategoryRoutes(group *echo.Group) {
	timeProvider := time_provider.NewTimeProvider(time.Now)

	// repositories
	categoryRepository := category_repository.NewRepository(server.db.GetInstance())

	// services
	createCategoryService := create_category.NewService(categoryRepository, timeProvider)
	getCategoryService := get_category.NewService(categoryRepository)
	getCategoriesService := get_categories.NewService(categoryRepository)
	deleteCategoryService := delete_category.NewService(categoryRepository)

	// handlers
	createCategoryHandler := create_category_handler.NewHandler(createCategoryService, getCategoriesService)
	getCategoryHandler := get_category_handler.NewHandler(getCategoryService)
	getCategoriesHandler := get_categories_handler.NewHandler(getCategoriesService)
	deleteCategoryHandler := delete_category_handler.NewHandler(deleteCategoryService)

	// routes
	group.POST("/categories", createCategoryHandler.Handle)
	group.GET("/categories", getCategoriesHandler.Handle)
	group.GET("/categories/:id", getCategoryHandler.Handle)
	group.DELETE("/categories/:id", deleteCategoryHandler.Handle)
}

func (server *Server) registerProductRoutes(group *echo.Group) {
	timeProvider := time_provider.NewTimeProvider(time.Now)

	// repositories
	productRepository := product_repository.NewRepository(server.db.GetInstance())
	categoryRepository := category_repository.NewRepository(server.db.GetInstance())

	// services
	createProductService := create_product.NewService(productRepository, categoryRepository, timeProvider)
	getProductService := get_product.NewService(productRepository)
	getProductsService := get_products.NewService(productRepository)
	deleteProductService := delete_product.NewService(productRepository)

	// handlers
	createProductHandler := create_product_handler.NewHandler(createProductService, getProductsService)
	getProductHandler := get_product_handler.NewHandler(getProductService)
	getProductsHandler := get_products_handler.NewHandler(getProductsService)
	deleteProductHandler := delete_product_handler.NewHandler(deleteProductService)

	// routes
	group.POST("/products", createProductHandler.Handle)
	group.GET("/products", getProductsHandler.Handle)
	group.GET("/products/:id", getProductHandler.Handle)
	group.DELETE("/products/:id", deleteProductHandler.Handle)
}
