package api

import (
	"bytes"
	"context"
	"errors"
	"firmguard/internal/model"
	"firmguard/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) CreateScan(ctx context.Context, scan *model.FirmwareScan) (*model.FirmwareScan, error) {
	args := m.Called(ctx, scan)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.FirmwareScan), args.Error(1)
}

func TestCreateScan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e := echo.New()
		svc := new(MockService)
		h := NewFirmwareScanHandler(svc)

		payload := `{"device_id":"d1", "firmware_version":"v1", "binary_hash":"h1"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/firmware-scans", bytes.NewBufferString(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		scan := &model.FirmwareScan{DeviceID: "d1", FirmwareVersion: "v1", BinaryHash: "h1"}
		svc.On("CreateScan", mock.Anything, mock.MatchedBy(func(s *model.FirmwareScan) bool {
			return s.DeviceID == "d1" && s.FirmwareVersion == "v1" && s.BinaryHash == "h1"
		})).Return(scan, nil)

		if assert.NoError(t, h.CreateScan(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		e := echo.New()
		h := NewFirmwareScanHandler(nil)
		req := httptest.NewRequest(http.MethodPost, "/v1/firmware-scans", bytes.NewBufferString("invalid"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		assert.NoError(t, h.CreateScan(c))
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("missing fields", func(t *testing.T) {
		e := echo.New()
		h := NewFirmwareScanHandler(nil)
		payload := `{"device_id":"d1"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/firmware-scans", bytes.NewBufferString(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		assert.NoError(t, h.CreateScan(c))
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("conflict", func(t *testing.T) {
		e := echo.New()
		svc := new(MockService)
		h := NewFirmwareScanHandler(svc)

		payload := `{"device_id":"d1", "firmware_version":"v1", "binary_hash":"h1"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/firmware-scans", bytes.NewBufferString(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		existing := &model.FirmwareScan{ID: 1, DeviceID: "d1", FirmwareVersion: "v1", BinaryHash: "h1"}
		svc.On("CreateScan", mock.Anything, mock.Anything).Return(existing, service.ErrScanAlreadyExists)

		assert.NoError(t, h.CreateScan(c))
		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		e := echo.New()
		svc := new(MockService)
		h := NewFirmwareScanHandler(svc)

		payload := `{"device_id":"d1", "firmware_version":"v1", "binary_hash":"h1"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/firmware-scans", bytes.NewBufferString(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		svc.On("CreateScan", mock.Anything, mock.Anything).Return(nil, errors.New("error"))

		assert.NoError(t, h.CreateScan(c))
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
