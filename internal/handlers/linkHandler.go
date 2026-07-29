package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"shortner/internal/service"
)

var _ http.Handler = (*LinkHandler)(nil) /* "проверка интерфейса во время компиляции" - создает
переменную и игнорирует ее. Переменная имеет тип интерфейса, на
соответствие которому мы хотим проверить нашу структуру.
В эту переменную кладем пустое значение типа "указатель типа
структуры". Это не создает объект -> не потребляет ресурсы,
но если интерфейс реализован неправильно, компиляция упадет
в этой строке*/

type LinkHandler struct {
	Service *service.ShortnerService
	logger  *slog.Logger
}

type ShortenLinkRequest struct {
	Link string `json:"link"`
}

type ShortenLinkResponse struct {
	Link string `json:"shortLink"`
}

func NewLinkHandler(s *service.ShortnerService, logger *slog.Logger) *LinkHandler {
	componentLogger := logger.With(slog.String("component", "link-handler"))
	return &LinkHandler{Service: s, logger: componentLogger}
}

func (h *LinkHandler) Handler(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("request received", slog.String("request_method", r.Method),
		slog.String("scheme", r.URL.Scheme),
		slog.String("host", r.Host),
		slog.String("url", r.URL.String()))
	switch r.Method {
	case http.MethodGet:
		h.redirect(w, r)
	case http.MethodPost:
		h.createShortLink(w, r)
	default:
		http.Error(w, fmt.Sprintf("invalid method: %s", r.Method), http.StatusMethodNotAllowed)
		h.logger.Info("invalid method", slog.String("method", r.Method))
	}
}

func (h *LinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Handler(w, r)
}

func (h *LinkHandler) createShortLink(w http.ResponseWriter, r *http.Request) {
	var data []byte
	data, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Info("unable to read request body", slog.String("error", err.Error()))
		http.Error(w, "Cannot read request", http.StatusBadRequest)
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Warn("unable to close request body", slog.String("error", err.Error()))
		}
	}()
	var shortenRequest ShortenLinkRequest
	if err := json.Unmarshal(data, &shortenRequest); err != nil {
		h.logger.Info("unable to parse request body", slog.String("error", err.Error()))
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := validateLink(shortenRequest.Link); err != nil {
		h.logger.Info("invalid link", slog.String("error", err.Error()))
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	shortLink, err := h.Service.ShortenLink(shortenRequest.Link)
	if err != nil {
		h.logger.Error("unable to shorten link", slog.String("error", err.Error()))
		http.Error(w, "Cannot shorten", http.StatusBadRequest)
		return
	}

	resp := ShortenLinkResponse{
		Link: shortLink,
	}

	respData, meer := json.Marshal(resp)
	if meer != nil {
		h.logger.Info("unable to marshal response", slog.String("error", meer.Error()))
		http.Error(w, "Cannot marshall", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, werr := w.Write(respData)
	if werr != nil {
		h.logger.Info("unable to write response", slog.String("error", werr.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *LinkHandler) redirect(w http.ResponseWriter, r *http.Request) {
	shortLink := r.URL.Path[1:]

	originalLink, err := h.Service.GetOriginalLink(shortLink)
	if err != nil {
		h.logger.Info("unable to get original link", slog.String("error", err.Error()))
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}

	fmt.Printf("Redirect %s to %s", shortLink, originalLink)
	fmt.Println()
	http.Redirect(w, r, originalLink, 301)
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

func (e ValidationError) Error() string {
	if e.error == nil {
		return fmt.Sprintf("validation error in %s: %s %v", e.value, e.message,
			e.error)
	}
	return fmt.Sprintf("validation error in %s: %s", e.value, e.message)
}
