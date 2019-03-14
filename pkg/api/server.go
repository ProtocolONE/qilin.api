package api

import (
	"encoding/base64"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/api/game"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/sys"
	"qilin-api/pkg/utils"
	"strconv"
)

type ServerOptions struct {
	ServerConfig     *conf.ServerConfig
	Jwt              *conf.Jwt
	Database         *orm.Database
	Mailer           sys.Mailer
	Notifier         sys.Notifier
	CentrifugoSecret string
}

type Server struct {
	db               *orm.Database
	echo             *echo.Echo
	serverConfig     *conf.ServerConfig
	notifier         sys.Notifier
	centrifugoSecret string

	Router      *echo.Group
	AdminRouter *echo.Group
	AuthRouter  *echo.Group
}

type QilinValidator struct {
	validator *validator.Validate
}

func (cv *QilinValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func NewServer(opts *ServerOptions) (*Server, error) {
	server := &Server{
		echo:             echo.New(),
		serverConfig:     opts.ServerConfig,
		db:               opts.Database,
		notifier:         opts.Notifier,
		centrifugoSecret: opts.CentrifugoSecret,
	}

	server.echo.HideBanner = true
	server.echo.HidePort = true
	server.echo.Debug = opts.ServerConfig.Debug

	server.echo.Use(ZapLogger(zap.L())) // logs all http requests
	server.echo.HTTPErrorHandler = server.QilinErrorHandler

	validate := validator.New()
	if err := utils.RegisterCustomValidations(validate); err != nil {
		return nil, err
	}
	validate.RegisterStructValidation(RatingStructLevelValidation, RatingsDTO{})

	server.echo.Validator = &QilinValidator{validator: validate}

	server.echo.Use(middleware.Recover())
	server.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		ExposeHeaders:    []string{"x-centrifugo-token", "x-items-count"},
		AllowHeaders:     []string{"authorization", "content-type"},
		AllowOrigins:     opts.ServerConfig.AllowOrigins,
		AllowCredentials: opts.ServerConfig.AllowCredentials,
	}))
	server.echo.Pre(middleware.RemoveTrailingSlash())

	server.Router = server.echo.Group("/api/v1")
	server.AdminRouter = server.echo.Group("/admin/api/v1")

	pemKey, err := base64.StdEncoding.DecodeString(opts.Jwt.SignatureSecret)
	if err != nil {
		return nil, errors.Wrap(err, "Decode JWT failed")
	}

	server.AdminRouter.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		ContextKey:    context.TokenKey,
		AuthScheme:    "Bearer",
		TokenLookup:   "header:Authorization",
		SigningKey:    pemKey,
		SigningMethod: opts.Jwt.Algorithm,
	}))

	server.Router.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		ContextKey:    context.TokenKey,
		AuthScheme:    "Bearer",
		TokenLookup:   "header:Authorization",
		SigningKey:    pemKey,
		SigningMethod: opts.Jwt.Algorithm,
	}))
	server.AuthRouter = server.echo.Group("/auth-api")

	if err := server.setupRoutes(opts.Jwt, opts.Mailer); err != nil {
		zap.L().Fatal("Fail to setup routes", zap.Error(err))
	}

	return server, nil
}

func (s *Server) Start() error {
	zap.L().Info("Starting http server", zap.Int("port", s.serverConfig.Port))

	return s.echo.Start(":" + strconv.Itoa(s.serverConfig.Port))
}

func (s *Server) setupRoutes(jwtConf *conf.Jwt, mailer sys.Mailer) error {
	userService, err := orm.NewUserService(s.db, jwtConf, mailer)
	if err != nil {
		return err
	}

	if err := InitUserRoutes(s, userService); err != nil {
		return err
	}

	vendorService, err := orm.NewVendorService(s.db)
	if err != nil {
		return err
	}

	if err := InitVendorRoutes(s, vendorService); err != nil {
		return err
	}

	mediaService, err := orm.NewMediaService(s.db)
	if err != nil {
		return err
	}

	if _, err := InitMediaRouter(s.Router, mediaService); err != nil {
		return err
	}

	gameService, err := orm.NewGameService(s.db)
	if err != nil {
		return err
	}

	priceService, err := orm.NewPriceService(s.db)
	if err != nil {
		return err
	}

	if _, err := InitPriceRouter(s.Router, priceService); err != nil {
		return err
	}

	ratingService, err := orm.NewRatingService(s.db)
	if err != nil {
		return err
	}

	if _, err := InitRatingsRouter(s.Router, ratingService); err != nil {
		return err
	}

	discountService, err := orm.NewDiscountService(s.db)
	if err != nil {
		return err
	}

	if _, err := InitDiscountsRouter(s.Router, discountService); err != nil {
		return err
	}

	clientOnboarding, err := orm.NewOnboardingService(s.db)
	if err != nil {
		return err
	}

	notificationService, err := orm.NewNotificationService(s.db, s.notifier, s.centrifugoSecret)

	if err != nil {
		return err
	}

	if _, err := InitClientOnboardingRouter(s.Router, clientOnboarding, notificationService); err != nil {
		return err
	}

	adminClientOnboarding, err := orm.NewAdminOnboardingService(s.db)
	if err != nil {
		return err
	}

	if _, err := InitAdminOnboardingRouter(s.AdminRouter, adminClientOnboarding, notificationService); err != nil {
		return err
	}

	if _, err := game.InitRoutes(s.Router, gameService); err != nil {
		return err
	}

	return nil
}
