package api

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/mongo"
	"strconv"
)

type ServerConfig struct {
	ServerConfig *conf.ServerConfig
	Log          *logrus.Entry
	Jwt          *conf.Jwt
	Session      *mongo.Session
}

type Server struct {
	log          *logrus.Entry
	session      *mongo.Session
	echo         *echo.Echo
	serverConfig *conf.ServerConfig
	Router       *echo.Group
}

func NewServer(config *ServerConfig) (*Server, error) {
	server := &Server{
		log:          config.Log,
		echo:         echo.New(),
		serverConfig: config.ServerConfig,
		session:      config.Session,
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
	gameService, err := mongo.NewGameService(s.session)
	if err != nil {
		return err
	}

	if err := InitGameRoutes(s, gameService); err != nil {
		return err
	}

	return nil
}
