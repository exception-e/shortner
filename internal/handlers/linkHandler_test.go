package handlers

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"shortner/internal/domain"
	"shortner/internal/service"
	logger "shortner/internal/tests"
	service_mock "shortner/mocks/service"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLinkHandler_CreateShortLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		prepare    func(mockService *service_mock.MockLinkService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			body: `{"link":"https://google.com"}`,
			prepare: func(mockService *service_mock.MockLinkService) {
				mockService.EXPECT().
					ShortenLink(gomock.Any(), "https://google.com").Return("http://localhost:8080/"+"3XqGtZ", nil).
					Times(1)
			},
			wantStatus: http.StatusOK,
			wantBody:   "3XqGtZ",
		},
		{
			name: "requestDecodeFails",
			body: ``,
			prepare: func(mockService *service_mock.MockLinkService) {
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "",
		},
		{
			name: "validationFails",
			body: `{"link":"https://[google.com]"}`,
			prepare: func(mockService *service_mock.MockLinkService) {
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "",
		},
		{
			name: "emptyURLFails",
			body: `{"link":""}`,
			prepare: func(mockService *service_mock.MockLinkService) {
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "",
		},
		{
			name: "alreadyExists",
			body: `{"link":"https://google.com"}`,
			prepare: func(mockService *service_mock.MockLinkService) {
				mockService.EXPECT().ShortenLink(gomock.Any(), "https://google.com").Return("", service.ErrAlreadyExists)
			},
			wantStatus: http.StatusConflict,
			wantBody:   "",
		},
		{
			name: "serviceFails",
			body: `{"link":"https://google.com"}`,
			prepare: func(mockService *service_mock.MockLinkService) {
				mockService.EXPECT().ShortenLink(gomock.Any(), "https://google.com").Return("", service.ErrNotSaved)
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockService := service_mock.NewMockLinkService(ctrl)
			tt.prepare(mockService)
			h := &LinkHandler{
				Service: mockService,
				logger:  slog.New(logger.NewTestHandler(t)),
			}
			req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.CreateShortLink(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestLinkHandler_Redirect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		alias        string
		prepare      func(mockService *service_mock.MockLinkService)
		useRouter    bool
		wantStatus   int
		wantLocation string
	}{
		{
			name:  "success",
			alias: "3XqGtZ",
			prepare: func(mockService *service_mock.MockLinkService) {
				mockService.EXPECT().
					GetOriginalLink(gomock.Any(), "3XqGtZ").
					Return(&domain.Link{OriginalURL: "https://google.com", Alias: "3XqGtZ"}, nil).
					Times(1)
			},
			useRouter:    true,
			wantStatus:   http.StatusMovedPermanently,
			wantLocation: "https://google.com",
		},
		{
			name:  "emptyAliasFails",
			alias: "",
			prepare: func(mockService *service_mock.MockLinkService) {
			},
			useRouter:    false,
			wantStatus:   http.StatusBadRequest,
			wantLocation: "",
		},
		{
			name:  "aliasNotFound",
			alias: "nonexisting",
			prepare: func(mockService *service_mock.MockLinkService) {
				mockService.EXPECT().
					GetOriginalLink(gomock.Any(), "nonexisting").
					Return(nil, service.ErrNotFound).
					Times(1)
			},
			useRouter:    true,
			wantStatus:   http.StatusNotFound,
			wantLocation: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockService := service_mock.NewMockLinkService(ctrl)
			tt.prepare(mockService)
			h := &LinkHandler{
				Service: mockService,
				logger:  slog.New(logger.NewTestHandler(t)),
			}
			req := httptest.NewRequest(http.MethodGet, "/"+tt.alias, nil)
			w := httptest.NewRecorder()
			if tt.useRouter {
				r := chi.NewRouter()
				r.Get("/{alias}", h.Redirect)
				r.ServeHTTP(w, req)
			} else {
				h.Redirect(w, req)
			}
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantLocation != "" {
				assert.Equal(t, tt.wantLocation, w.Header().Get("Location"))
			}
		})
	}
}

func Test_validateLink(t *testing.T) {
	type args struct {
		link string
	}
	tests := []struct {
		name    string
		link    string
		wantErr bool
	}{
		{"valid https", "https://google.com", false},
		{"valid http", "http://google.com", false},
		{"empty", "", true},
		{"invalid format", "not-a-url", true},
		{"unsupported scheme", "ftp://google.com", true},
		{"no host", "https://", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLink(tt.link)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
