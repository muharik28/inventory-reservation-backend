package test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"github.com/muharik28/inventory-reservation-backend/api/rest/api/routes"
	"github.com/muharik28/inventory-reservation-backend/configs"
	infoHandler "github.com/muharik28/inventory-reservation-backend/internal/application/handlers/info/impl"
	invHandler "github.com/muharik28/inventory-reservation-backend/internal/application/handlers/inventory/impl"
	resHandler "github.com/muharik28/inventory-reservation-backend/internal/application/handlers/reservation/impl"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/request"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/response"
	infoRepo "github.com/muharik28/inventory-reservation-backend/internal/application/repositories/info/impl"
	invRepo "github.com/muharik28/inventory-reservation-backend/internal/application/repositories/inventory/impl"
	resRepo "github.com/muharik28/inventory-reservation-backend/internal/application/repositories/reservation/impl"
	infoUseCase "github.com/muharik28/inventory-reservation-backend/internal/application/usecases/info/impl"
	invUseCase "github.com/muharik28/inventory-reservation-backend/internal/application/usecases/inventory/impl"
	"github.com/muharik28/inventory-reservation-backend/internal/application/usecases/reservation"
	resUseCase "github.com/muharik28/inventory-reservation-backend/internal/application/usecases/reservation/impl"
	domainInv "github.com/muharik28/inventory-reservation-backend/internal/domain/entities/inventory"
	domainRes "github.com/muharik28/inventory-reservation-backend/internal/domain/entities/reservation"
	dbsql "github.com/muharik28/inventory-reservation-backend/internal/pkg/database/sql"
)

// The integration tests run against a real PostgreSQL instance (no mocks, no
// in-memory substitutes). Each test gets a dedicated throwaway database that
// is recreated from migrations/20261010_init.sql, so tests never interfere
// with the dev/demo database or with a concurrently running scheduler.
//
// Start PostgreSQL before running:
//
//	docker compose -f deployments/docker-compose.yml up -d postgres
const testDBName = "inventory_test_db"

type testEnv struct {
	engine        *gin.Engine
	db            *bun.DB
	reservationUC reservation.ReservationUseCase
	cfg           *configs.Config
}

type errorResponse struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Error  string `json:"error"`
}

// -----------------------------------------------------------------------------
// Test environment helpers
// -----------------------------------------------------------------------------

func adminDSN(cfg *configs.Config, dbName string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password,
		cfg.Database.Host, cfg.Database.Port,
		dbName, cfg.Database.SSLMode,
	)
}

func pingPostgres(t *testing.T, cfg *configs.Config) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(adminDSN(cfg, "postgres"))))
	defer sqldb.Close()

	if err := sqldb.PingContext(ctx); err != nil {
		t.Skipf("PostgreSQL not reachable, start it with: docker compose -f deployments/docker-compose.yml up -d postgres (error: %v)", err)
	}
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	cfg := configs.LoadConfig()
	cfg.AppEnv = "test"
	cfg.Database.DBName = testDBName
	cfg.ReservationTTL = 5 * time.Minute
	cfg.SchedulerInterval = 10 * time.Second

	pingPostgres(t, cfg)

	// Recreate the throwaway test database for hermetic isolation.
	admin := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(adminDSN(cfg, "postgres"))))
	_, err := admin.ExecContext(context.Background(), "DROP DATABASE IF EXISTS "+testDBName+" WITH (FORCE)")
	if err != nil {
		t.Fatalf("failed to drop test database: %v", err)
	}
	if _, err := admin.ExecContext(context.Background(), "CREATE DATABASE "+testDBName); err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	if err := admin.Close(); err != nil {
		t.Fatalf("failed to close admin connection: %v", err)
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(adminDSN(cfg, testDBName))))
	db := bun.NewDB(sqldb, pgdialect.New())
	t.Cleanup(func() {
		_ = db.Close()
	})

	// Apply the committed migration file so the schema under test is the
	// real one shipped in the repository.
	applyMigrationFile(t, db, "../migrations/20261010_init.sql")

	// Wire the real handler -> use case -> repository stack against the DB.
	txMgr := dbsql.NewBunTxManager(db)
	iRepo := infoRepo.NewInfoRepository(db, cfg)
	inventoryRepo := invRepo.NewInventoryRepository(db)
	reservationRepo := resRepo.NewReservationRepository(db)

	iUseCase := infoUseCase.NewInfoUseCase(iRepo)
	inventoryUseCase := invUseCase.NewInventoryUseCase(txMgr, inventoryRepo, reservationRepo)
	reservationUseCase := resUseCase.NewReservationUseCase(txMgr, inventoryRepo, reservationRepo, cfg)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	routes.SetupRouter(engine, routes.RouterParams{
		InfoHandler:        infoHandler.NewInfoHandler(iUseCase),
		InventoryHandler:   invHandler.NewInventoryHandler(inventoryUseCase),
		ReservationHandler: resHandler.NewReservationHandler(reservationUseCase),
	})

	return &testEnv{engine: engine, db: db, reservationUC: reservationUseCase, cfg: cfg}
}

func applyMigrationFile(t *testing.T, db *bun.DB, path string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read migration file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Strip full-line comments before splitting on ';' so comment text can
	// never be executed as a statement.
	var lines []string
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		lines = append(lines, line)
	}

	for _, stmt := range strings.Split(strings.Join(lines, "\n"), ";") {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("migration statement failed (%s): %v", firstLine(stmt), err)
		}
	}
}

func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
			return trimmed
		}
	}
	return ""
}

// -----------------------------------------------------------------------------
// HTTP helpers
// -----------------------------------------------------------------------------

func doJSON(e *gin.Engine, method, target string, body any) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, target, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	return w
}

func getStock(t *testing.T, e *testEnv, itemID string) response.StockResponse {
	t.Helper()

	w := doJSON(e.engine, http.MethodGet, "/api/v1/inventory/stock?item_id="+itemID, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("GET stock %s: expected 200, got %d: %s", itemID, w.Code, w.Body.String())
	}

	var res response.StockResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode stock response: %v", err)
	}
	return res
}

func createItem(t *testing.T, e *testEnv, itemID string, totalStock int) {
	t.Helper()

	w := doJSON(e.engine, http.MethodPost, "/api/v1/inventory/items", request.CreateItemRequest{
		ItemID:     itemID,
		TotalStock: totalStock,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create item %s: expected 201, got %d: %s", itemID, w.Code, w.Body.String())
	}
}

func reserve(t *testing.T, e *testEnv, userID, itemID string, qty int) *httptest.ResponseRecorder {
	t.Helper()

	return doJSON(e.engine, http.MethodPost, "/api/v1/inventory/reserve", request.ReserveStockRequest{
		UserID:   userID,
		ItemID:   itemID,
		Quantity: qty,
	})
}

func confirm(t *testing.T, e *testEnv, reservationID string) *httptest.ResponseRecorder {
	t.Helper()

	return doJSON(e.engine, http.MethodPost, "/api/v1/inventory/confirm", request.ConfirmReservationRequest{
		ReservationID: reservationID,
	})
}

func uniqueItemID() string {
	return "item_test_" + strings.ReplaceAll(uuid.NewString()[:8], "-", "")
}

func forceExpire(t *testing.T, e *testEnv, reservationID string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := e.db.ExecContext(ctx,
		"UPDATE reservations SET expires_at = NOW() - INTERVAL '1 second' WHERE reservation_id = ?",
		reservationID); err != nil {
		t.Fatalf("failed to force expiry: %v", err)
	}
}

func reservationStatus(t *testing.T, e *testEnv, reservationID string) domainRes.Status {
	t.Helper()

	var status string
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := e.db.NewSelect().
		TableExpr("reservations").
		ColumnExpr("status").
		Where("reservation_id = ?", reservationID).
		Scan(ctx, &status)
	if err != nil {
		t.Fatalf("failed to read reservation status: %v", err)
	}
	return domainRes.Status(status)
}

// -----------------------------------------------------------------------------
// Unit tests (pure domain logic, no database required)
// -----------------------------------------------------------------------------

func TestDomainLogic(t *testing.T) {
	inv := &domainInv.Inventory{ItemID: "item_4021", TotalStock: 100}

	if got := inv.AvailableStock(15); got != 85 {
		t.Fatalf("expected available stock 85, got %d", got)
	}

	if !inv.CanReserve(85, 15) {
		t.Fatalf("expected CanReserve(85, 15) to be true")
	}

	if inv.CanReserve(86, 15) {
		t.Fatalf("expected CanReserve(86, 15) to be false")
	}

	if inv.CanReserve(0, 0) || inv.CanReserve(-1, 0) {
		t.Fatalf("expected non-positive quantities to be rejected")
	}

	if got := inv.AvailableStock(120); got != 0 {
		t.Fatalf("expected available stock clamped to 0, got %d", got)
	}

	now := time.Now()
	res := &domainRes.Reservation{
		Status:    domainRes.StatusPending,
		ExpiresAt: now.Add(-1 * time.Minute),
	}
	if !res.IsExpired(now) {
		t.Fatalf("expected reservation to be expired")
	}

	active := &domainRes.Reservation{
		Status:    domainRes.StatusPending,
		ExpiresAt: now.Add(5 * time.Minute),
	}
	if active.IsExpired(now) {
		t.Fatalf("expected reservation to be active")
	}

	confirmed := &domainRes.Reservation{Status: domainRes.StatusConfirmed}
	if confirmed.IsPending() || confirmed.IsExpired(now) {
		t.Fatalf("confirmed reservation must not be pending or expirable")
	}
}

// -----------------------------------------------------------------------------
// Integration tests: GET /api/v1/inventory/stock
// -----------------------------------------------------------------------------

func TestStockEndpoint(t *testing.T) {
	e := newTestEnv(t)

	// Seeded by the migration: item_4021 has 100 total and 15 active
	// reserved units.
	stock := getStock(t, e, "item_4021")
	if stock.TotalStock != 100 || stock.ReservedStock != 15 || stock.AvailableStock != 85 {
		t.Fatalf("unexpected seeded stock: %+v", stock)
	}

	// Missing item_id parameter.
	w := doJSON(e.engine, http.MethodGet, "/api/v1/inventory/stock", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing item_id, got %d", w.Code)
	}
	assertErrorCode(t, w, "INVALID_REQUEST")

	// Unknown item.
	w = doJSON(e.engine, http.MethodGet, "/api/v1/inventory/stock?item_id=does_not_exist", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown item, got %d", w.Code)
	}
	assertErrorCode(t, w, "ITEM_NOT_FOUND")
}

// -----------------------------------------------------------------------------
// Integration tests: POST /api/v1/inventory/reserve
// -----------------------------------------------------------------------------

func TestReserveEndpointSuccessAndErrors(t *testing.T) {
	e := newTestEnv(t)

	// Success against the seeded item (85 available).
	before := time.Now()
	w := reserve(t, e, "USR-1", "item_4021", 5)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var res response.ReserveStockResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode reserve response: %v", err)
	}
	if !strings.HasPrefix(res.ReservationID, "res_") || res.ReservationID == "res_" {
		t.Fatalf("reservation id must be non-empty with res_ prefix, got %q", res.ReservationID)
	}
	if res.ItemID != "item_4021" || res.Quantity != 5 {
		t.Fatalf("unexpected reservation payload: %+v", res)
	}

	ttl := res.ExpiresAt.Sub(before)
	if ttl < 4*time.Minute+30*time.Second || ttl > 5*time.Minute+30*time.Second {
		t.Fatalf("expires_at should be ~5 minutes in the future, got %v", ttl)
	}

	// Stock reflects the new reservation immediately.
	stock := getStock(t, e, "item_4021")
	if stock.ReservedStock != 20 || stock.AvailableStock != 80 {
		t.Fatalf("expected reserved=20 available=80, got %+v", stock)
	}

	// Invalid JSON body.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/inventory/reserve", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "INVALID_REQUEST")

	// Zero quantity.
	w = reserve(t, e, "USR-1", "item_4021", 0)
	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "INVALID_QUANTITY")

	// Missing user_id.
	w = doJSON(e.engine, http.MethodPost, "/api/v1/inventory/reserve", request.ReserveStockRequest{
		ItemID: "item_4021", Quantity: 1,
	})
	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "INVALID_REQUEST")

	// Unknown item.
	w = reserve(t, e, "USR-1", "no_such_item", 1)
	assertStatus(t, w, http.StatusNotFound)
	assertErrorCode(t, w, "ITEM_NOT_FOUND")

	// Insufficient stock.
	w = reserve(t, e, "USR-1", "item_4021", 1000)
	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "INSUFFICIENT_STOCK")
}

// -----------------------------------------------------------------------------
// Integration tests: POST /api/v1/inventory/confirm
// -----------------------------------------------------------------------------

func TestConfirmEndpointSuccessAndErrors(t *testing.T) {
	e := newTestEnv(t)

	item := uniqueItemID()
	createItem(t, e, item, 10)

	// Happy path: confirming decrements physical stock exactly once.
	w := reserve(t, e, "USR-2", item, 4)
	res := decodeReserveResponse(t, w)

	w = confirm(t, e, res.ReservationID)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on confirm, got %d: %s", w.Code, w.Body.String())
	}

	stock := getStock(t, e, item)
	if stock.TotalStock != 6 || stock.ReservedStock != 0 || stock.AvailableStock != 6 {
		t.Fatalf("expected total=6 reserved=0 after confirm, got %+v", stock)
	}

	// Confirming twice must fail with a clear error.
	w = confirm(t, e, res.ReservationID)
	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "RESERVATION_ALREADY_HANDLED")

	// Unknown reservation.
	w = confirm(t, e, "res_does_not_exist")
	assertStatus(t, w, http.StatusNotFound)
	assertErrorCode(t, w, "RESERVATION_NOT_FOUND")

	// Expired reservation: stock must not be reduced.
	w = reserve(t, e, "USR-2", item, 2)
	expired := decodeReserveResponse(t, w)
	forceExpire(t, e, expired.ReservationID)

	w = confirm(t, e, expired.ReservationID)
	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "RESERVATION_EXPIRED")

	stock = getStock(t, e, item)
	if stock.TotalStock != 6 {
		t.Fatalf("expired confirm must not reduce physical stock, got %+v", stock)
	}
}

// -----------------------------------------------------------------------------
// Integration test: expiry correctness (PostgreSQL is the source of truth)
// -----------------------------------------------------------------------------

func TestExpiryReleasesStockWithoutWorker(t *testing.T) {
	e := newTestEnv(t)

	item := uniqueItemID()
	createItem(t, e, item, 10)

	w := reserve(t, e, "USR-3", item, 3)
	res := decodeReserveResponse(t, w)

	stock := getStock(t, e, item)
	if stock.ReservedStock != 3 || stock.AvailableStock != 7 {
		t.Fatalf("expected reserved=3 available=7, got %+v", stock)
	}

	// Simulate the TTL elapsing without waiting 5 minutes. Availability is
	// derived from expires_at, so the reservation must stop counting even
	// though the cleanup worker has not run yet.
	forceExpire(t, e, res.ReservationID)

	stock = getStock(t, e, item)
	if stock.ReservedStock != 0 || stock.AvailableStock != 10 {
		t.Fatalf("expired reservation must not reduce available stock, got %+v", stock)
	}

	// The worker then performs the bookkeeping status flip.
	count, err := e.reservationUC.ReleaseExpiredReservations(context.Background())
	if err != nil {
		t.Fatalf("release worker failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 released reservation, got %d", count)
	}
	if got := reservationStatus(t, e, res.ReservationID); got != domainRes.StatusExpired {
		t.Fatalf("expected reservation status EXPIRED, got %s", got)
	}

	// The full quantity can be reserved again.
	w = reserve(t, e, "USR-3", item, 10)
	if w.Code != http.StatusOK {
		t.Fatalf("expected to reserve full stock after expiry, got %d: %s", w.Code, w.Body.String())
	}
}

// -----------------------------------------------------------------------------
// Integration test: database constraints prevent invalid stored quantities
// -----------------------------------------------------------------------------

func TestDatabaseConstraintsPreventInvalidData(t *testing.T) {
	e := newTestEnv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Negative reservation quantity must be rejected by CHECK (quantity > 0).
	_, err := e.db.ExecContext(ctx, `
		INSERT INTO reservations (reservation_id, user_id, item_id, quantity, status, expires_at)
		VALUES ('res_bad_qty', 'USR-X', 'item_4021', -1, 'PENDING', NOW() + INTERVAL '5 minutes')`)
	if err == nil {
		t.Fatalf("expected CHECK constraint to reject negative quantity")
	}

	// Negative stock must be rejected by CHECK (total_stock >= 0).
	_, err = e.db.ExecContext(ctx, `UPDATE inventories SET total_stock = -5 WHERE item_id = 'item_4021'`)
	if err == nil {
		t.Fatalf("expected CHECK constraint to reject negative total_stock")
	}
}

// -----------------------------------------------------------------------------
// Concurrency / stress tests
// -----------------------------------------------------------------------------

func TestConcurrentReservationsDoNotOversell(t *testing.T) {
	e := newTestEnv(t)

	const totalStock = 10
	const workers = 50

	item := uniqueItemID()
	createItem(t, e, item, totalStock)

	statuses := make(chan int, workers)
	reservationIDs := make(chan string, workers)
	reservationIDsSet := make(map[string]struct{})
	var mu sync.Mutex

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			w := reserve(t, e, fmt.Sprintf("USR-C-%d", i), item, 1)
			statuses <- w.Code

			if w.Code == http.StatusOK {
				var res response.ReserveStockResponse
				_ = json.Unmarshal(w.Body.Bytes(), &res)
				mu.Lock()
				reservationIDsSet[res.ReservationID] = struct{}{}
				mu.Unlock()
				reservationIDs <- res.ReservationID
			}
		}(i)
	}
	wg.Wait()
	close(statuses)
	close(reservationIDs)

	success, insufficient, other := 0, 0, 0
	for code := range statuses {
		switch code {
		case http.StatusOK:
			success++
		case http.StatusBadRequest:
			insufficient++
		default:
			other++
		}
	}

	if success != totalStock {
		t.Fatalf("expected exactly %d successful reservations, got %d", totalStock, success)
	}
	if insufficient != workers-totalStock {
		t.Fatalf("expected %d insufficient-stock rejections, got %d", workers-totalStock, insufficient)
	}
	if other != 0 {
		t.Fatalf("unexpected non-200/non-400 responses: %d", other)
	}
	if len(reservationIDsSet) != success {
		t.Fatalf("reservation IDs must be unique, got %d unique for %d successes", len(reservationIDsSet), success)
	}

	// No overselling: reserved equals stock, available is zero.
	stock := getStock(t, e, item)
	if stock.ReservedStock != totalStock || stock.AvailableStock != 0 || stock.TotalStock != totalStock {
		t.Fatalf("expected reserved=10 available=0 total=10, got %+v", stock)
	}

	// Confirming every successful reservation decrements physical stock to
	// exactly zero, never below.
	for reservationID := range reservationIDs {
		w := confirm(t, e, reservationID)
		if w.Code != http.StatusOK {
			t.Fatalf("confirm %s failed: %d %s", reservationID, w.Code, w.Body.String())
		}
	}

	stock = getStock(t, e, item)
	if stock.TotalStock != 0 || stock.ReservedStock != 0 || stock.AvailableStock != 0 {
		t.Fatalf("expected total=0 reserved=0 after all confirms, got %+v", stock)
	}
}

func TestConcurrentConfirmExactlyOnce(t *testing.T) {
	e := newTestEnv(t)

	item := uniqueItemID()
	createItem(t, e, item, 20)

	w := reserve(t, e, "USR-D", item, 5)
	res := decodeReserveResponse(t, w)

	const workers = 12
	statuses := make(chan int, workers)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			statuses <- confirm(t, e, res.ReservationID).Code
		}()
	}
	wg.Wait()
	close(statuses)

	confirmed, rejected := 0, 0
	for code := range statuses {
		switch code {
		case http.StatusOK:
			confirmed++
		case http.StatusBadRequest:
			rejected++
		default:
			t.Fatalf("unexpected confirm status code %d", code)
		}
	}

	if confirmed != 1 {
		t.Fatalf("reservation must be confirmed exactly once, got %d confirmations", confirmed)
	}
	if rejected != workers-1 {
		t.Fatalf("expected %d already-handled rejections, got %d", workers-1, rejected)
	}

	stock := getStock(t, e, item)
	if stock.TotalStock != 15 || stock.ReservedStock != 0 || stock.AvailableStock != 15 {
		t.Fatalf("physical stock must decrease exactly once, got %+v", stock)
	}
}

// -----------------------------------------------------------------------------
// Assertion helpers
// -----------------------------------------------------------------------------

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Fatalf("expected status %d, got %d: %s", expected, w.Code, w.Body.String())
	}
}

func assertErrorCode(t *testing.T, w *httptest.ResponseRecorder, expectedCode string) {
	t.Helper()

	var res errorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode error response: %v (%s)", err, w.Body.String())
	}
	if res.Code != expectedCode {
		t.Fatalf("expected error code %s, got %q (%s)", expectedCode, res.Code, w.Body.String())
	}
	if res.Status != "error" {
		t.Fatalf("expected status field 'error', got %q", res.Status)
	}
}

func decodeReserveResponse(t *testing.T, w *httptest.ResponseRecorder) response.ReserveStockResponse {
	t.Helper()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 reserve response, got %d: %s", w.Code, w.Body.String())
	}

	var res response.ReserveStockResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode reserve response: %v", err)
	}
	return res
}
