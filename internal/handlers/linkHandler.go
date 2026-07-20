package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"shortner/internal/service"
)

var _ http.Handler = (*LinkHandler)(nil)

type LinkHandler struct {
	Service *service.ShorterService
}

type ShortenLinkRequest struct {
	Link string `json:"link"`
}

type ShortenLinkResponse struct {
	Link string `json:"shortLink"`
}

func NewLinkHandler(s *service.ShorterService) *LinkHandler {
	return &LinkHandler{Service: s}
}

func (h *LinkHandler) Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.redirect(w, r)
	case http.MethodPost:
		h.createShortLink(w, r)
	}
}

func (h *LinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Handler(w, r)
}

func (h *LinkHandler) createShortLink(w http.ResponseWriter, r *http.Request) {
	var data []byte
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Cannot read request: %v", err)
		return
	}
	defer r.Body.Close()

	var shortenRequest ShortenLinkRequest
	jsonErr := json.Unmarshal(data, &shortenRequest)
	if jsonErr != nil {
		log.Printf("Cannot unmarshall: %v", jsonErr)
	}

	if err := service.ValidateLink(shortenRequest.Link); err != nil {
		log.Printf("Entered line is not an URL (%s): %v", shortenRequest.Link, err)
	}

	shortLink := h.Service.ShortenLink(shortenRequest.Link)

	resp := ShortenLinkResponse{
		Link: shortLink,
	}

	respData, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	_, werr := w.Write(respData)
	if werr != nil {
		fmt.Printf("Response error %v", werr)
	}
}

func (h *LinkHandler) redirect(w http.ResponseWriter, r *http.Request) {
	shortLink := r.URL.Path[1:]

	originalLink, err := h.Service.GetOriginalLink(shortLink)
	if err != nil {
		fmt.Printf("Link not found: %v", err)
		return
	}

	fmt.Printf("Original link: %s. ", originalLink)
	fmt.Printf("Redirect to %s", originalLink)
	http.Redirect(w, r, originalLink, 301)
}
