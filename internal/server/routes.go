package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	e.GET("/", s.HelloWorldHandler)

	e.GET("/health", s.healthHandler)

	v1 := e.Group("/v1")
	v1.POST("/firmware-scans", s.scanHdl.CreateScan)

	findings := v1.Group("/findings")
	findings.PATCH("/vulns", s.vulnHdl.RegisterVulnerabilities)
	findings.GET("/vulns", s.vulnHdl.GetVulnerabilities)

	return e
}
func (s *Server) HelloWorldHandler(c echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c echo.Context) error {
	stats, err := s.db.Health()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, stats)
	}
	return c.JSON(http.StatusOK, stats)
}
