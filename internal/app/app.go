package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/config"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/handler"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/platform/database"
	cachedrepo "github.com/Group-project-bridgeon/Catalog-Service/internal/repository/cache"
	postgresrepo "github.com/Group-project-bridgeon/Catalog-Service/internal/repository/postgres"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository/postgres/catalogdb"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/router"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func Run(log zerolog.Logger) error {
	if err := godotenv.Load(); err != nil &&
		!errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf(
			"load .env: %w",
			err,
		)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf(
			"load config: %w",
			err,
		)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	db, err := database.NewPool(
		ctx,
		cfg.DatabaseURL,
		database.PoolConfig{
			MaxConns:          cfg.Database.MaxConns,
			MinConns:          cfg.Database.MinConns,
			MaxConnLifetime:   cfg.Database.MaxConnLifetime,
			MaxConnIdleTime:   cfg.Database.MaxConnIdleTime,
			HealthCheckPeriod: cfg.Database.HealthCheckPeriod,
			StatementTimeout:  cfg.Database.StatementTimeout,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"connect to database: %w",
			err,
		)
	}
	defer db.Close()

	queries := catalogdb.New(db)

	categoryRepository := cachedrepo.NewCategoryRepository(
		postgresrepo.NewCategoryRepository(
			queries,
		),
		cfg.Cache.ReferenceTTL,
	)

	categoryService := service.NewCategoryService(
		categoryRepository,
	)

	categoryHandler := handler.NewCategoryHandler(
		categoryService,
		log,
	)

	brandRepository := cachedrepo.NewBrandRepository(
		postgresrepo.NewBrandRepository(
			queries,
		),
		cfg.Cache.ReferenceTTL,
	)

	brandService := service.NewBrandService(
		brandRepository,
	)

	brandHandler := handler.NewBrandHandler(
		brandService,
		log,
	)

	unitRepository := cachedrepo.NewUnitRepository(
		postgresrepo.NewUnitRepository(
			queries,
		),
		cfg.Cache.ReferenceTTL,
	)

	unitService := service.NewUnitService(
		unitRepository,
	)

	unitHandler := handler.NewUnitHandler(
		unitService,
		log,
	)

	productRepository := postgresrepo.NewProductRepository(
		queries,
	)

	productService := service.NewProductService(
		productRepository,
	)

	productHandler := handler.NewProductHandler(
		productService,
		log,
	)

	productVariantRepository := postgresrepo.NewProductVariantRepository(
		queries,
	)

	productVariantService := service.NewProductVariantService(
		productVariantRepository,
	)

	productVariantHandler := handler.NewProductVariantHandler(
		productVariantService,
		log,
	)

	attributeRepository := cachedrepo.NewAttributeRepository(
		postgresrepo.NewAttributeRepository(
			queries,
		),
		cfg.Cache.ReferenceTTL,
	)

	attributeService := service.NewAttributeService(
		attributeRepository,
	)

	attributeHandler := handler.NewAttributeHandler(
		attributeService,
		log,
	)

	attributeValueRepository := postgresrepo.NewAttributeValueRepository(
		queries,
	)

	attributeValueService := service.NewAttributeValueService(
		attributeValueRepository,
	)

	attributeValueHandler := handler.NewAttributeValueHandler(
		attributeValueService,
		log,
	)

	variantValueRepository := postgresrepo.NewVariantValueRepository(
		queries,
	)

	variantValueService := service.NewVariantValueService(
		variantValueRepository,
	)

	variantValueHandler := handler.NewVariantValueHandler(
		variantValueService,
		log,
	)

	productImageRepository := postgresrepo.NewProductImageRepository(
		queries,
	)

	productImageService := service.NewProductImageService(
		productImageRepository,
	)

	productImageHandler := handler.NewProductImageHandler(
		productImageService,
		log,
	)

	priceListRepository := postgresrepo.NewPriceListRepository(
		queries,
	)

	priceListService := service.NewPriceListService(
		priceListRepository,
	)

	priceListHandler := handler.NewPriceListHandler(
		priceListService,
		log,
	)

	priceRepository := postgresrepo.NewPriceRepository(
		queries,
	)

	priceService := service.NewPriceService(
		priceRepository,
	)

	priceHandler := handler.NewPriceHandler(
		priceService,
		log,
	)

	productDetailRepository := postgresrepo.NewProductDetailRepository(
		queries,
	)

	productDetailService := service.NewProductDetailService(
		productDetailRepository,
	)

	productDetailHandler := handler.NewProductDetailHandler(
		productDetailService,
		log,
	)

	gin.SetMode(cfg.GinMode)

	httpHandler := router.New(
		db,
		log,
		cfg.HTTP.MaxRequestBodyBytes,
		cfg.HTTP.SlowRequestThreshold,
	)

	router.RegisterCategoryRoutes(
		httpHandler,
		categoryHandler,
	)

	router.RegisterBrandRoutes(
		httpHandler,
		brandHandler,
	)

	router.RegisterUnitRoutes(
		httpHandler,
		unitHandler,
	)

	router.RegisterProductRoutes(
		httpHandler,
		productHandler,
	)

	router.RegisterProductVariantRoutes(
		httpHandler,
		productVariantHandler,
	)

	router.RegisterAttributeRoutes(
		httpHandler,
		attributeHandler,
	)

	router.RegisterAttributeValueRoutes(
		httpHandler,
		attributeValueHandler,
	)

	router.RegisterVariantValueRoutes(
		httpHandler,
		variantValueHandler,
	)

	router.RegisterProductImageRoutes(
		httpHandler,
		productImageHandler,
	)

	router.RegisterPriceListRoutes(
		httpHandler,
		priceListHandler,
	)

	router.RegisterPriceRoutes(
		httpHandler,
		priceHandler,
	)

	router.RegisterProductDetailRoutes(
		httpHandler,
		productDetailHandler,
	)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpHandler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	serverErr := make(chan error, 1)

	go func() {
		log.Info().
			Str("port", cfg.Port).
			Msg("catalog service starting")

		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf(
				"serve http: %w",
				err,
			)
		}

		return nil

	case <-ctx.Done():
		log.Info().
			Msg("catalog service shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		shutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf(
			"shutdown http server: %w",
			err,
		)
	}

	log.Info().
		Msg("catalog service stopped")

	return nil
}
