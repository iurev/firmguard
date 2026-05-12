package api

import (
	"firmguard/internal/model"
	"firmguard/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type FirmwareScanHandler struct {
	svc service.FirmwareScanService
}

func NewFirmwareScanHandler(svc service.FirmwareScanService) *FirmwareScanHandler {
	return &FirmwareScanHandler{svc: svc}
}

func (h *FirmwareScanHandler) CreateScan(c echo.Context) error {
	var scan model.FirmwareScan
	if err := c.Bind(&scan); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}

	if scan.DeviceID == "" || scan.FirmwareVersion == "" || scan.BinaryHash == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "missing required fields"})
	}

	result, err := h.svc.CreateScan(c.Request().Context(), &scan)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create scan"})
	}

	return c.JSON(http.StatusAccepted, result)
}
