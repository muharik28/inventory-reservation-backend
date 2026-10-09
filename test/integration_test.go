package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muharik28/inventory-reservation-backend/api/rest/api/routes"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/request"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/response"
	domainInv "github.com/muharik28/inventory-reservation-backend/internal/domain/entities/inventory"
	domainRes "github.com/muharik28/inventory-reservation-backend/internal/domain/entities/reservation"
)

// Mock Inventory UseCase
type mockInventoryUseCase struct{}

func (m *mockInventoryUseCase) GetStock(ctx context.Context, itemID string) (*response.StockResponse, error) {
	if itemID == "item_4021" {
		return &response.StockResponse{
			ItemID:         "item_4021",
			TotalStock:     100,
			ReservedStock:  15,
			AvailableStock: 85,
		}, nil
	}
	return nil, domainInv.ErrInventoryNotFound
}

func (m *mockInventoryUseCase) CreateOrUpdateItem(ctx context.Context, req request.CreateItemRequest) (*response.StockResponse, error) {
	return &response.StockResponse{
		ItemID:         req.ItemID,
		TotalStock:     req.TotalStock,
		ReservedStock:  0,
		AvailableStock: req.TotalStock,
	}, nil
}

// Mock Reservation UseCase
type mockReservationUseCase struct{}

func (m *mockReservationUseCase) ReserveStock(ctx context.Context, req request.ReserveStockRequest) (*response.ReserveStockResponse, error) {
	return &response.ReserveStockResponse{
		Status:        "success",
		ReservationID: "res_883291",
		ItemID:        req.ItemID,
		Quantity:      req.Quantity,
		ExpiresAt:     time.Now().Add(5 * time.Minute),
	}, nil
}

func (m *mockReservationUseCase) ConfirmReservation(ctx context.Context, req request.ConfirmReservationRequest) (*response.ConfirmReservationResponse, error) {
	return &response.ConfirmReservationResponse{
		Status:        "success",
		ReservationID: req.ReservationID,
		ConfirmedAt:   time.Now(),
	}, nil
}

func (m *mockReservationUseCase) ReleaseExpiredReservations(ctx context.Context) (int, error) {
	return 1, nil
}

// Mock Info UseCase
type mockInfoUseCase struct{}

func (m *mockInfoUseCase) GetInfo(ctx context.Context) (*response.InfoDetailResponse, error) {
	return &response.InfoDetailResponse{
		ServiceName: "inventory-reservation-backend",
		Version:     "1.0.0",
		Environment: "test",
		Status:      "healthy",
		DBStatus:    "connected",
		ServerTime:  time.Now(),
	}, nil
}

func TestDomainLogic(t *testing.T) {
	inv := &domainInv.Inventory{
		ItemID:        "item_4021",
		TotalStock:    100,
		ReservedStock: 15,
	}

	if inv.AvailableStock() != 85 {
		t.Fatalf("expected available stock 85, got %d", inv.AvailableStock())
	}

	if !inv.CanReserve(85) {
		t.Fatalf("expected CanReserve(85) to be true")
	}

	if inv.CanReserve(86) {
		t.Fatalf("expected CanReserve(86) to be false")
	}

	res := &domainRes.Reservation{
		Status:    domainRes.StatusPending,
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}

	if !res.IsExpired(time.Now()) {
		t.Fatalf("expected reservation to be expired")
	}
}

func TestRoutesAndHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	infoH := &mockInfoHandler{}
	invH := &mockInvHandler{}
	resH := &mockResHandler{}

	routes.SetupRouter(engine, routes.RouterParams{
		InfoHandler:        infoH,
		InventoryHandler:   invH,
		ReservationHandler: resH,
	})

	// Test GET /health
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for /health, got %d", w.Code)
	}

	// Test GET /api/v1/inventory/stock?item_id=item_4021
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/inventory/stock?item_id=item_4021", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for /api/v1/inventory/stock, got %d", w.Code)
	}
}

type mockInfoHandler struct{}

func (h *mockInfoHandler) GetInfo(c *gin.Context)    { c.JSON(http.StatusOK, gin.H{"status": "ok"}) }
func (h *mockInfoHandler) HealthCheck(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "healthy"}) }

type mockInvHandler struct{}

func (h *mockInvHandler) GetStock(c *gin.Context)   { c.JSON(http.StatusOK, gin.H{"item_id": "item_4021"}) }
func (h *mockInvHandler) CreateItem(c *gin.Context) { c.JSON(http.StatusCreated, gin.H{"status": "created"}) }

type mockResHandler struct{}

func (h *mockResHandler) Reserve(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "success"}) }
func (h *mockResHandler) Confirm(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "success"}) }
