package api

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/orm"
	"strconv"
)

type ServerConfig struct {
	ServerConfig *conf.ServerConfig
	Log          *logrus.Entry
	Jwt          *conf.Jwt
	Database     *orm.Database
}

type Server struct {
	log          *logrus.Entry
	db           *orm.Database
	echo         *echo.Echo
	serverConfig *conf.ServerConfig
	Router       *echo.Group
}

func NewServer(config *ServerConfig) (*Server, error) {
	server := &Server{
		log:          config.Log,
		echo:         echo.New(),
		serverConfig: config.ServerConfig,
		db:           config.Database,
	}

	server.echo.Logger = Logger{config.Log.Logger}
	server.echo.Use(LoggerHandler)

	server.echo.Use(middleware.Logger())
	server.echo.Use(middleware.Recover())
	server.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowHeaders: []string{"authorization", "content-type"},
	}))

	server.Router = server.echo.Group("/api/v1")
	/*
		server.Router.Use(middleware.JWTWithConfig(middleware.JWTConfig{
			SigningKey:    config.Jwt.SignatureSecret,
			SigningMethod: config.Jwt.Algorithm,
		}))
	*/

	if err := server.setupRoutes(); err != nil {
		server.log.Fatal(err)
	}

	return server, nil
}

func (s *Server) Start() error {
	return s.echo.Start(":" + strconv.Itoa(s.serverConfig.Port))
}

func (s *Server) setupRoutes() error {
	gameService, err := orm.NewGameService(s.db)
	if err != nil {
		return err
	}

	if err := InitGameRoutes(s, gameService); err != nil {
		return err
	}

	userService, err := orm.NewUserService(s.db)
	if err != nil {
		return err
	}

	if err := InitUserRoutes(s, userService); err != nil {
		return err
	}

	return nil
}
