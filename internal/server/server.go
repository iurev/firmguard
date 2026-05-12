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

	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
	"github.com/riverqueue/river"
)

type Server struct {
	port int

	db      database.Service
	scanHdl *api.FirmwareScanHandler
	vulnHdl *api.VulnerabilityHandler
}

func NewServer(db database.Service, riverClient *river.Client[pgx.Tx]) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	scanRepo := repository.NewFirmwareScanRepository(db.GetPool())
	scanSvc := service.NewFirmwareScanService(db, scanRepo, riverClient)
	scanHdl := api.NewFirmwareScanHandler(scanSvc)

	vulnRepo := repository.NewVulnerabilityRepository(db.GetPool())
	vulnSvc := service.NewVulnerabilityService(vulnRepo)
	vulnHdl := api.NewVulnerabilityHandler(vulnSvc)

	srv := &Server{
		port: port,

		db:      db,
		scanHdl: scanHdl,
		vulnHdl: vulnHdl,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", srv.port),
		Handler:      srv.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
