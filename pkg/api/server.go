package api

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/orm"
	"strconv"
)

type ServerOptions struct {
	ServerConfig *conf.ServerConfig
	Log          *logrus.Entry
	Jwt          *conf.Jwt
	Database     *orm.Database
}

type Server struct {
	log          	*logrus.Entry
	db           	*orm.Database
	echo         	*echo.Echo
	serverConfig 	*conf.ServerConfig

	Router       	*echo.Group
	AuthRouter   	*echo.Group
}

func NewServer(opts *ServerOptions) (*Server, error) {
	server := &Server{
		log:          opts.Log,
		echo:         echo.New(),
		serverConfig: opts.ServerConfig,
		db:           opts.Database,
	}

	server.echo.Logger = Logger{opts.Log.Logger}
	server.echo.Use(LoggerHandler)

	server.echo.Use(middleware.Logger())
	server.echo.Use(middleware.Recover())
	server.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowHeaders: []string{"authorization", "content-type"},
		AllowOrigins: opts.ServerConfig.AllowOrigins,
		AllowCredentials: opts.ServerConfig.AllowCredentials,
	}))
	server.echo.Pre(middleware.RemoveTrailingSlash())

	server.Router = server.echo.Group("/api/v1")
	server.Router.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		TokenLookup: 	"cookie:token",
		SigningKey:    	opts.Jwt.SignatureSecret,
		SigningMethod: 	opts.Jwt.Algorithm,
	}))
	server.AuthRouter = server.echo.Group("/auth-api")

	if err := server.setupRoutes(opts.Jwt); err != nil {
		server.log.Fatal(err)
	}

	return server, nil
}

func (s *Server) Start() error {
	return s.echo.Start(":" + strconv.Itoa(s.serverConfig.Port))
}

func (s *Server) setupRoutes(jwtConf *conf.Jwt) error {
	gameService, err := orm.NewGameService(s.db)
	if err != nil {
		return err
	}

	if err := InitGameRoutes(s, gameService); err != nil {
		return err
	}

	userService, err := orm.NewUserService(s.db, jwtConf)
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

	return nil
}
