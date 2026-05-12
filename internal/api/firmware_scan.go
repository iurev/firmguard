package api

import (
	"errors"
	"firmguard/internal/model"
	"firmguard/internal/service"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

type FirmwareScanHandler struct {
	svc service.FirmwareScanService
}

func NewFirmwareScanHandler(svc service.FirmwareScanService) *FirmwareScanHandler {
	return &FirmwareScanHandler{svc: svc}
}

func (h *FirmwareScanHandler) CreateScan(c echo.Context) error {
	// NOTE: FirmwareScan domain model is used directly as the request type. Fields like
	// ID, Status, Vulns, and timestamps sent by the client are ignored/overwritten by the service.
	var scan model.FirmwareScan
	if err := c.Bind(&scan); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}

	if scan.DeviceID == "" || scan.FirmwareVersion == "" || scan.BinaryHash == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "missing required fields"})
	}

	if len(scan.DeviceID) > 255 || len(scan.FirmwareVersion) > 100 || len(scan.BinaryHash) > 64 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "field exceeds maximum length"})
	}

	result, isNew, err := h.svc.CreateScan(c.Request().Context(), &scan)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create scan"})
	}

	if isNew {
		return c.JSON(http.StatusAccepted, result)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *FirmwareScanHandler) GetScan(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid scan ID"})
	}

	scan, err := h.svc.GetScan(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "scan not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to get scan"})
	}

	return c.JSON(http.StatusOK, scan)
}
