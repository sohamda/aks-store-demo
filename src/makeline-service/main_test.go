package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// MockOrderRepo is a mock implementation of the OrderRepo interface for testing.
type MockOrderRepo struct {
	pendingOrders    []Order
	getPendingErr    error
	getOrderFunc     func(id string) (Order, error)
	insertOrdersFunc func(orders []Order) error
	updateOrderFunc  func(order Order) error
}

func (m *MockOrderRepo) GetPendingOrders() ([]Order, error) {
	return m.pendingOrders, m.getPendingErr
}

func (m *MockOrderRepo) GetOrder(id string) (Order, error) {
	if m.getOrderFunc != nil {
		return m.getOrderFunc(id)
	}
	return Order{}, errors.New("not found")
}

func (m *MockOrderRepo) InsertOrders(orders []Order) error {
	if m.insertOrdersFunc != nil {
		return m.insertOrdersFunc(orders)
	}
	return nil
}

func (m *MockOrderRepo) UpdateOrder(order Order) error {
	if m.updateOrderFunc != nil {
		return m.updateOrderFunc(order)
	}
	return nil
}

// setupTestRouter creates a Gin router in test mode with the provided OrderRepo.
func setupTestRouter(repo OrderRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	service := NewOrderService(repo)
	router.Use(OrderMiddleware(service))
	router.GET("/order/:id", getOrder)
	router.PUT("/order", updateOrder)
	return router
}

// TestUnmarshalOrderFromQueue tests the unmarshalOrderFromQueue function.
func TestUnmarshalOrderFromQueue(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		wantStatus Status
	}{
		{
			name:       "valid order with items",
			input:      `{"customerId":"cust-1","items":[{"productId":1,"quantity":2,"price":9.99}]}`,
			wantErr:    false,
			wantStatus: Pending,
		},
		{
			name:       "valid order with no items",
			input:      `{"customerId":"cust-2","items":[]}`,
			wantErr:    false,
			wantStatus: Pending,
		},
		{
			name:    "invalid json",
			input:   `not valid json`,
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := unmarshalOrderFromQueue([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if order.Status != tt.wantStatus {
				t.Errorf("expected status %v, got %v", tt.wantStatus, order.Status)
			}
			if order.OrderID == "" {
				t.Error("expected non-empty OrderID after unmarshal")
			}
		})
	}
}

// TestGetOrder tests the getOrder HTTP handler.
func TestGetOrder(t *testing.T) {
	tests := []struct {
		name         string
		orderID      string
		getOrderFunc func(id string) (Order, error)
		wantStatus   int
	}{
		{
			name:    "valid order",
			orderID: "42",
			getOrderFunc: func(id string) (Order, error) {
				return Order{OrderID: id, CustomerID: "cust-1", Status: Pending}, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-numeric order id returns bad request",
			orderID:    "abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "database error returns internal server error",
			orderID: "99",
			getOrderFunc: func(id string) (Order, error) {
				return Order{}, errors.New("db unavailable")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockOrderRepo{getOrderFunc: tt.getOrderFunc}
			router := setupTestRouter(mock)

			req := httptest.NewRequest(http.MethodGet, "/order/"+tt.orderID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected HTTP %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

// TestGetOrderResponseBody verifies that the returned JSON body matches the stored order.
func TestGetOrderResponseBody(t *testing.T) {
	expected := Order{OrderID: "7", CustomerID: "cust-99", Status: Processing, Items: []Item{{Product: 3, Quantity: 1, Price: 4.99}}}

	mock := &MockOrderRepo{
		getOrderFunc: func(id string) (Order, error) {
			return expected, nil
		},
	}
	router := setupTestRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/order/7", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var got Order
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.CustomerID != expected.CustomerID {
		t.Errorf("expected CustomerID %q, got %q", expected.CustomerID, got.CustomerID)
	}
	if got.Status != expected.Status {
		t.Errorf("expected Status %v, got %v", expected.Status, got.Status)
	}
}

// TestUpdateOrder tests the updateOrder HTTP handler.
func TestUpdateOrder(t *testing.T) {
	tests := []struct {
		name            string
		order           Order
		updateOrderFunc func(order Order) error
		wantStatus      int
	}{
		{
			name:  "valid update returns 200",
			order: Order{OrderID: "5", CustomerID: "cust-1", Status: Complete},
			updateOrderFunc: func(order Order) error {
				return nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-numeric order id returns bad request",
			order:      Order{OrderID: "not-a-number", CustomerID: "cust-1", Status: Complete},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "database error returns internal server error",
			order: Order{OrderID: "5", CustomerID: "cust-1", Status: Processing},
			updateOrderFunc: func(order Order) error {
				return errors.New("db write failed")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockOrderRepo{updateOrderFunc: tt.updateOrderFunc}
			router := setupTestRouter(mock)

			body, err := json.Marshal(tt.order)
			if err != nil {
				t.Fatalf("failed to marshal order: %v", err)
			}
			req := httptest.NewRequest(http.MethodPut, "/order", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected HTTP %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

// TestUpdateOrderInvalidBody tests that a malformed request body is rejected.
func TestUpdateOrderInvalidBody(t *testing.T) {
	mock := &MockOrderRepo{}
	router := setupTestRouter(mock)

	req := httptest.NewRequest(http.MethodPut, "/order", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError && w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 or 500 for invalid body, got %d", w.Code)
	}
}

// TestOrderMiddleware verifies that the middleware correctly injects OrderService into the context.
func TestOrderMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &MockOrderRepo{}
	service := NewOrderService(mock)

	var capturedService *OrderService

	router := gin.New()
	router.Use(OrderMiddleware(service))
	router.GET("/test", func(c *gin.Context) {
		s, ok := c.MustGet("orderService").(*OrderService)
		if !ok {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		capturedService = s
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if capturedService != service {
		t.Error("middleware did not inject the expected OrderService")
	}
}
