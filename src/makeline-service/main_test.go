package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockOrderRepo struct {
	getOrderFn    func(id string) (Order, error)
	updateOrderFn func(order Order) error
}

func (m mockOrderRepo) GetPendingOrders() ([]Order, error) {
	return nil, nil
}

func (m mockOrderRepo) GetOrder(id string) (Order, error) {
	if m.getOrderFn != nil {
		return m.getOrderFn(id)
	}
	return Order{}, nil
}

func (m mockOrderRepo) InsertOrders(orders []Order) error {
	return nil
}

func (m mockOrderRepo) UpdateOrder(order Order) error {
	if m.updateOrderFn != nil {
		return m.updateOrderFn(order)
	}
	return nil
}

func TestOrderMiddlewareSetsOrderService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mockOrderRepo{}
	service := NewOrderService(repo)
	router := gin.New()
	router.Use(OrderMiddleware(service))
	router.GET("/test", func(c *gin.Context) {
		value, ok := c.Get("orderService")
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		retrieved, typeOK := value.(*OrderService)
		if !typeOK || retrieved != service {
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Status(http.StatusOK)
	})

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
}

func TestGetOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name           string
		pathID         string
		repo           mockOrderRepo
		expectedStatus int
		expectedRepoID string
	}{
		{
			name:           "returns bad request for non-numeric id",
			pathID:         "abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "returns internal server error when repo fails",
			pathID: "42",
			repo: mockOrderRepo{getOrderFn: func(id string) (Order, error) {
				return Order{}, errors.New("repo unavailable")
			}},
			expectedStatus: http.StatusInternalServerError,
			expectedRepoID: "42",
		},
		{
			name:   "returns order when lookup succeeds",
			pathID: "0007",
			repo: mockOrderRepo{getOrderFn: func(id string) (Order, error) {
				return Order{OrderID: id, CustomerID: "customer-1", Status: Pending}, nil
			}},
			expectedStatus: http.StatusOK,
			expectedRepoID: "7",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			calledRepoID := ""
			repo := tc.repo
			if tc.expectedRepoID != "" {
				original := repo.getOrderFn
				repo.getOrderFn = func(id string) (Order, error) {
					calledRepoID = id
					if original != nil {
						return original(id)
					}
					return Order{}, nil
				}
			}

			service := NewOrderService(repo)
			router := gin.New()
			router.Use(OrderMiddleware(service))
			router.GET("/order/:id", getOrder)

			resp := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/order/"+tc.pathID, nil)
			router.ServeHTTP(resp, req)

			if resp.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, resp.Code)
			}

			if tc.expectedRepoID != "" && calledRepoID != tc.expectedRepoID {
				t.Fatalf("expected repo id %q, got %q", tc.expectedRepoID, calledRepoID)
			}
		})
	}
}

func TestUpdateOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name           string
		body           string
		repo           mockOrderRepo
		expectedStatus int
		expectedRepoID string
	}{
		{
			name:           "returns bad request for malformed json",
			body:           `{"orderId":`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "returns bad request for non-numeric order id",
			body:           `{"orderId":"not-a-number","customerId":"1","items":[],"status":0}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "returns internal server error when repository update fails",
			body: `{"orderId":"17","customerId":"1","items":[],"status":0}`,
			repo: mockOrderRepo{updateOrderFn: func(order Order) error {
				return errors.New("update failed")
			}},
			expectedStatus: http.StatusInternalServerError,
			expectedRepoID: "17",
		},
		{
			name: "returns success response when repository update succeeds",
			body: `{"orderId":"0012","customerId":"1","items":[],"status":2}`,
			repo: mockOrderRepo{updateOrderFn: func(order Order) error {
				return nil
			}},
			expectedStatus: http.StatusOK,
			expectedRepoID: "12",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			calledRepoID := ""
			repo := tc.repo
			if tc.expectedRepoID != "" {
				original := repo.updateOrderFn
				repo.updateOrderFn = func(order Order) error {
					calledRepoID = order.OrderID
					if original != nil {
						return original(order)
					}
					return nil
				}
			}

			service := NewOrderService(repo)
			router := gin.New()
			router.Use(OrderMiddleware(service))
			router.PUT("/order", updateOrder)

			resp := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/order", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(resp, req)

			if resp.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, resp.Code)
			}

			if tc.expectedRepoID != "" && calledRepoID != tc.expectedRepoID {
				t.Fatalf("expected repo id %q, got %q", tc.expectedRepoID, calledRepoID)
			}
		})
	}
}

func TestUnmarshalOrderFromQueue(t *testing.T) {
	testCases := []struct {
		name        string
		payload     []byte
		wantErr     bool
		customerID  string
		itemsLength int
	}{
		{
			name:        "returns order with generated id and pending status",
			payload:     []byte(`{"customerId":"customer-123","items":[{"productId":1,"quantity":1,"price":9.99}],"status":2}`),
			wantErr:     false,
			customerID:  "customer-123",
			itemsLength: 1,
		},
		{
			name:    "returns error for invalid json",
			payload: []byte(`{"customerId":`),
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			order, err := unmarshalOrderFromQueue(tc.payload)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			if _, convErr := strconv.Atoi(order.OrderID); convErr != nil {
				t.Fatalf("expected numeric order id, got %q", order.OrderID)
			}

			if order.Status != Pending {
				t.Fatalf("expected status %v, got %v", Pending, order.Status)
			}

			if order.CustomerID != tc.customerID {
				t.Fatalf("expected customer id %q, got %q", tc.customerID, order.CustomerID)
			}

			if len(order.Items) != tc.itemsLength {
				t.Fatalf("expected %d items, got %d", tc.itemsLength, len(order.Items))
			}
		})
	}
}

func TestGetOrderReturnsJSONBodyOnSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := NewOrderService(mockOrderRepo{getOrderFn: func(id string) (Order, error) {
		return Order{OrderID: id, CustomerID: "c-1", Status: Processing}, nil
	}})

	router := gin.New()
	router.Use(OrderMiddleware(service))
	router.GET("/order/:id", getOrder)

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/order/8", nil)
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var responseOrder Order
	if err := json.Unmarshal(resp.Body.Bytes(), &responseOrder); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if responseOrder.OrderID != "8" {
		t.Fatalf("expected order id %q, got %q", "8", responseOrder.OrderID)
	}
}
