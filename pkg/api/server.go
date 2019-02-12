package api

import (
	"encoding/base64"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/api/game"
	"qilin-api/pkg/utils"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/sys"
	"strconv"
)

type ServerOptions struct {
	ServerConfig *conf.ServerConfig
	Log          *logrus.Entry
	Jwt          *conf.Jwt
	Database     *orm.Database
	Mailer       sys.Mailer
}

type Server struct {
	log          *logrus.Entry
	db           *orm.Database
	echo         *echo.Echo
	serverConfig *conf.ServerConfig

	Router     *echo.Group
	AuthRouter *echo.Group
}

type QilinValidator struct {
	validator *validator.Validate
}

func (cv *QilinValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func NewServer(opts *ServerOptions) (*Server, error) {
	server := &Server{
		log:          opts.Log,
		echo:         echo.New(),
		serverConfig: opts.ServerConfig,
		db:           opts.Database,
	}

	server.echo.Debug = opts.ServerConfig.Debug
	server.echo.Logger = Logger{opts.Log.Logger}
	server.echo.Use(LoggerHandler) // logs all http requests
	server.echo.HTTPErrorHandler = server.QilinErrorHandler

	validate := validator.New()
	if err := utils.RegisterCustomValidations(validate); err != nil {
		return nil, err
	}
	validate.RegisterStructValidation(RatingStructLevelValidation, RatingsDTO{})

	server.echo.Validator = &QilinValidator{validator: validate}

	server.echo.Use(middleware.Recover())
	server.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowHeaders:     []string{"authorization", "content-type"},
		AllowOrigins:     opts.ServerConfig.AllowOrigins,
		AllowCredentials: opts.ServerConfig.AllowCredentials,
	}))
	server.echo.Pre(middleware.RemoveTrailingSlash())

	server.Router = server.echo.Group("/api/v1")

	pemKey, err := base64.StdEncoding.DecodeString(opts.Jwt.SignatureSecret)
	if err != nil {
		return nil, errors.Wrap(err, "Decode JWT failed")
	}

	server.Router.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		ContextKey:    context.TokenKey,
		AuthScheme:    "Bearer",
		TokenLookup:   "header:Authorization",
		SigningKey:    pemKey,
		SigningMethod: opts.Jwt.Algorithm,
	}))
	server.AuthRouter = server.echo.Group("/auth-api")

	if err := server.setupRoutes(opts.Jwt, opts.Mailer); err != nil {
		server.log.Fatal(err)
	}

	return server, nil
}

func (s *Server) Start() error {
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

	if _, err := game.InitRoutes(s.Router, gameService); err != nil {
		return err
	}

	return nil
}
