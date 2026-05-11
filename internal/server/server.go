package server

import (
	"firmguard/internal/api"
	"firmguard/internal/database"
	"firmguard/internal/repository"
	"firmguard/internal/service"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port int

	db      database.Service
	scanHdl *api.FirmwareScanHandler
	vulnHdl *api.VulnerabilityHandler
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	db := database.New()
	
	scanRepo := repository.NewFirmwareScanRepository(db.GetDB())
	scanSvc := service.NewFirmwareScanService(scanRepo)
	scanHdl := api.NewFirmwareScanHandler(scanSvc)

	vulnRepo := repository.NewVulnerabilityRepository(db.GetDB())
	vulnSvc := service.NewVulnerabilityService(vulnRepo)
	vulnHdl := api.NewVulnerabilityHandler(vulnSvc)

	NewServer := &Server{
		port: port,

		db:      db,
		scanHdl: scanHdl,
		vulnHdl: vulnHdl,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
