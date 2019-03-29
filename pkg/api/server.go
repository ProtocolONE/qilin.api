package api

import (
	"github.com/ProtocolONE/authone-jwt-verifier-golang"
	jwt_middleware "github.com/ProtocolONE/authone-jwt-verifier-golang/middleware/echo"
	"github.com/ProtocolONE/rbac"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	qilin_middleware "qilin-api/pkg/api/rbac_echo"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/sys"
	"qilin-api/pkg/utils"
	"strconv"
)

type ServerOptions struct {
	ServerConfig     *conf.ServerConfig
	Auth1            *conf.Auth1
	Database         *orm.Database
	Mailer           sys.Mailer
	Notifier         sys.Notifier
	CentrifugoSecret string
	Enforcer         *rbac.Enforcer
}

type Server struct {
	db               *orm.Database
	echo             *echo.Echo
	serverConfig     *conf.ServerConfig
	notifier         sys.Notifier
	centrifugoSecret string
	enforcer         *rbac.Enforcer

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
		enforcer:         opts.Enforcer,
	}

	server.echo.HideBanner = true
	server.echo.HidePort = true
	server.echo.Debug = opts.ServerConfig.Debug

	ownerProvider := orm.NewOwnerProvider(server.db)

	server.echo.Use(ZapLogger(zap.L())) // logs all http requests
	server.echo.Use(middleware.Recover())
	server.echo.Use(qilin_middleware.NewAppContextMiddleware(ownerProvider, server.enforcer))

	server.echo.HTTPErrorHandler = server.QilinErrorHandler

	validate := validator.New()
	if err := utils.RegisterCustomValidations(validate); err != nil {
		return nil, err
	}
	validate.RegisterStructValidation(RatingStructLevelValidation, RatingsDTO{})

	server.echo.Validator = &QilinValidator{validator: validate}

	server.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		ExposeHeaders:    []string{"x-centrifugo-token", "x-items-count"},
		AllowHeaders:     []string{"authorization", "content-type"},
		AllowOrigins:     opts.ServerConfig.AllowOrigins,
		AllowCredentials: opts.ServerConfig.AllowCredentials,
	}))
	server.echo.Pre(middleware.RemoveTrailingSlash())

	server.Router = server.echo.Group("/api/v1")
	server.AdminRouter = server.echo.Group("/admin/api/v1")

	settings := jwtverifier.Config{
		ClientID:     opts.Auth1.ClientId,
		ClientSecret: opts.Auth1.ClientSecret,
		Scopes:       []string{"openid", "offline"},
		RedirectURL:  "",
		Issuer:       opts.Auth1.Issuer,
	}
	jwtv := jwtverifier.NewJwtVerifier(settings)

	server.AdminRouter.Use(jwt_middleware.AuthOneJwtWithConfig(jwtv))
	server.Router.Use(jwt_middleware.AuthOneJwtWithConfig(jwtv))
	server.AuthRouter = server.echo.Group("/auth-api")

	if err := server.setupRoutes(ownerProvider, opts.Mailer, jwtv); err != nil {
		zap.L().Fatal("Fail to setup routes", zap.Error(err))
	}

	return server, nil
}

func (s *Server) Start() error {
	zap.L().Info("Starting http server", zap.Int("port", s.serverConfig.Port))

	return s.echo.Start(":" + strconv.Itoa(s.serverConfig.Port))
}

func (s *Server) setupRoutes(ownerProvider model.OwnerProvider, mailer sys.Mailer, verifier *jwtverifier.JwtVerifier) error {
	notificationService, err := orm.NewNotificationService(s.db, s.notifier, s.centrifugoSecret)
	if err != nil {
		return err
	}

	userService, err := orm.NewUserService(s.db, mailer)
	if err != nil {
		return err
	}
	if err := InitUserRoutes(s, userService, verifier); err != nil {
		return err
	}

	mediaService, err := orm.NewMediaService(s.db)
	if err != nil {
		return err
	}
	if _, err := InitMediaRouter(s.Router, mediaService); err != nil {
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
	if _, err := InitClientOnboardingRouter(s.Router, clientOnboarding, notificationService); err != nil {
		return err
	}

	membershipService := orm.NewMembershipService(s.db, ownerProvider, s.enforcer, mailer, "")
	if err := membershipService.Init(); err != nil {
		return err
	}

	if _, err := InitClientMembershipRouter(s.Router, membershipService); err != nil {
		return err
	}

	packageService, err := orm.NewPackageService(s.db)
	if err != nil {
		return err
	}
	if _, err := InitPackageRouter(s.Router, packageService); err != nil {
		return err
	}

	bundleService, err := orm.NewBundleService(s.db)
	if err != nil {
		return err
	}
	if _, err := InitBundleRouter(s.Router, bundleService); err != nil {
		return err
	}

	adminClientOnboarding, err := orm.NewAdminOnboardingService(s.db, membershipService, ownerProvider)
	if err != nil {
		return err
	}
	if _, err := InitAdminOnboardingRouter(s.AdminRouter, adminClientOnboarding, notificationService); err != nil {
		return err
	}

	gameService, err := orm.NewGameService(s.db)
	if err != nil {
		return err
	}

	vendorService, err := orm.NewVendorService(s.db, membershipService)
	if err != nil {
		return err
	}

	if _, err := InitRoutes(s.Router, gameService, userService); err != nil {
		return err
	}

	if err := InitVendorRoutes(s.Router, vendorService, userService); err != nil {
		return err
	}

	return nil
}
