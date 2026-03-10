package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// mockOrderRepo is a test double that implements OrderRepo
type mockOrderRepo struct {
	pingErr error
}

func (m *mockOrderRepo) GetPendingOrders() ([]Order, error) { return nil, nil }
func (m *mockOrderRepo) GetOrder(_ string) (Order, error)   { return Order{}, nil }
func (m *mockOrderRepo) InsertOrders(_ []Order) error       { return nil }
func (m *mockOrderRepo) UpdateOrder(_ Order) error          { return nil }
func (m *mockOrderRepo) Ping() error                        { return m.pingErr }

func setupRouter(repo OrderRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	orderService := NewOrderService(repo)
	router := gin.New()
	router.Use(OrderMiddleware(orderService))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": os.Getenv("APP_VERSION"),
		})
	})
	router.GET("/ready", readyHandler)
	return router
}

func TestHealthEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "returns 200 ok",
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"ok"`,
		},
	}

	router := setupRouter(&mockOrderRepo{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/health", nil)
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.expectedBody != "" && !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestReadyEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		pingErr        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "returns 200 when database is reachable",
			pingErr:        nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"ok"`,
		},
		{
			name:           "returns 503 when database is unreachable",
			pingErr:        errors.New("connection refused"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `"status":"not ready"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRouter(&mockOrderRepo{pingErr: tt.pingErr})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/ready", nil)
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.expectedBody != "" && !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}
