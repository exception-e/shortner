package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"shortner/internal/service"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//var _ http.Handler = (*LinkHandler)(nil) /* "Проверка интерфейса во время компиляции" - создает
//переменную и игнорирует ее. Переменная имеет тип интерфейса, на
//соответствие которому мы хотим проверить нашу структуру.
//В эту переменную кладем пустое значение типа "указатель типа
//структуры". Это не создает объект -> не потребляет ресурсы,
//но если интерфейс реализован неправильно, компиляция упадет
//в этой строке*/

type LinkHandler struct {
	Service service.LinkService
	logger  *slog.Logger
}

type ShortenLinkRequest struct {
	Link string `json:"link"`
}

type ShortenLinkResponse struct {
	ShortLink string `json:"shortLink"`
}

func NewLinkHandler(s service.LinkService, logger *slog.Logger) *LinkHandler {
	componentLogger := logger.With(slog.String("component", "link-handler"))
	return &LinkHandler{Service: s, logger: componentLogger}
}

func (h *LinkHandler) CreateShortLink(w http.ResponseWriter, r *http.Request) {

	var shortenRequest ShortenLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&shortenRequest); err != nil {
		h.respondError(r.Context(), w, http.StatusBadRequest, "Unable to parse request body", err)
		return
	}

	if err := validateLink(shortenRequest.Link); err != nil {
		h.respondError(r.Context(), w, http.StatusBadRequest, "Invalid link", err)
		return
	}

	alias, err := h.Service.ShortenLink(r.Context(), shortenRequest.Link)
	if err != nil { //TODO специфические ошибки
		if errors.Is(err, service.ErrAlreadyExists) {
			h.respondError(r.Context(), w, http.StatusConflict, "Unable to shorten link", err)
			return
		}
		h.respondError(r.Context(), w, http.StatusInternalServerError, "Unable to shorten link", err)
		return
	}

	resp := ShortenLinkResponse{
		ShortLink: alias,
	}

	err = h.respondJSON(r.Context(), w, http.StatusOK, resp)
	if err != nil {
		h.respondError(r.Context(), w, http.StatusInternalServerError, "Unable to write response", err)
		return
	}
}

func (h *LinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	alias := strings.TrimSpace(chi.URLParam(r, "alias"))
	if alias == "" {
		h.respondError(r.Context(), w, http.StatusBadRequest, "Missing alias",
			&ValidationError{message: "missing alias", value: ""})
		return
	}
	originalLink, err := h.Service.GetOriginalLink(r.Context(), alias)
	if err != nil {
		h.respondError(r.Context(), w, http.StatusNotFound, "Unable to get original link", err)
		return
	}
	h.logger.InfoContext(r.Context(), "redirect",
		"alias", alias,
		"original_url", originalLink.OriginalURL,
		"user_agent", r.UserAgent(),
		"ip", getClientIP(r))

	http.Redirect(w, r, originalLink.OriginalURL, http.StatusMovedPermanently)
}

func validateLink(link string) error {
	if link == "" {
		return &ValidationError{value: link, message: "empty link"}
	}
	parsed, err := url.Parse(link)
	if err != nil {
		return &ValidationError{value: link, message: "Invalid url format", error: err}
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return &ValidationError{value: link, message: "unsupported scheme:"}
	}
	if parsed.Host == "" {
		return &ValidationError{value: parsed.Host, message: "invalid host"}
	}
	return nil
}

type ValidationError struct {
	message string
	value   string
	error   error
}

func (e *ValidationError) Error() string {
	if e.error != nil {
		return fmt.Sprintf("validation error in %s: %s %v", e.value, e.message,
			e.error)
	}
	return fmt.Sprintf("validation error in %s: %s", e.value, e.message)
}

func (h *LinkHandler) respondError(ctx context.Context,
	w http.ResponseWriter,
	status int,
	message string,
	err error) {
	requestId, ok := ctx.Value(middleware.RequestIDKey).(string)
	if !ok || requestId == "" {
		requestId = "unknown"
	}
	h.logger.WarnContext(ctx, "request error",
		"status", status,
		"message", message,
		"error", err,
		"error_type", fmt.Sprintf("%T", err),
		"request_id", requestId)
	http.Error(w, message, status)
}

func (h *LinkHandler) respondJSON(ctx context.Context, w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.ErrorContext(ctx, "response encoding failed", "error", err)
		return err
	}
	return nil
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		//если несколько IP через запятую, берем первый
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	// проверяем X-Real-IP
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	// если нет прокси, берем RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
