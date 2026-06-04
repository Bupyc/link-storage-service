package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Bupyc/link-storage-service/internal/repository"
	"github.com/Bupyc/link-storage-service/internal/service"
)

type LinkHandler struct {
	service *service.LinkService
}

type createLinkRequest struct {
	URL string `json:"url"`
}

type createLinkResponse struct {
	ShortCode string `json:"short_code"`
}

type getLinkResponse struct {
	URL    string `json:"url"`
	Visits int64  `json:"visits"`
}

type getStatsResponse struct {
	ShortCode string `json:"short_code"`
	URL       string `json:"url"`
	Visits    int64  `json:"visits"`
	CreatedAt string `json:"created_at"`
}

type listLinksResponse struct {
	Links  []linkItemResponse `json:"links"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
	Total  int64              `json:"total"`
}

type linkItemResponse struct {
	ShortCode string `json:"short_code"`
	URL       string `json:"url"`
	Visits    int64  `json:"visits"`
	CreatedAt string `json:"created_at"`
}

func NewLinkHandler(service *service.LinkService) *LinkHandler {
	return &LinkHandler{
		service: service,
	}
}

func (h *LinkHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /links", h.createLink)
	mux.HandleFunc("GET /links", h.listLinks)
	mux.HandleFunc("GET /links/{short_code}", h.getLink)
	mux.HandleFunc("DELETE /links/{short_code}", h.deleteLink)
	mux.HandleFunc("GET /links/{short_code}/stats", h.getStats)
}

func (h *LinkHandler) createLink(w http.ResponseWriter, r *http.Request) {
	var req createLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	link, err := h.service.Create(r.Context(), req.URL)
	if err != nil {
		if err == service.ErrInvalidURL {
			writeError(w, http.StatusBadRequest, "invalid url")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, createLinkResponse{
		ShortCode: link.ID,
	})
}

func (h *LinkHandler) listLinks(w http.ResponseWriter, r *http.Request) {
	limit := 10
	offset := 0
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = parsedLimit
	}
	if rawOffset := r.URL.Query().Get("offset"); rawOffset != "" {
		parsedOffset, err := strconv.Atoi(rawOffset)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = parsedOffset
	}
	links, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list links")
		return
	}
	total, err := h.service.Count(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to count links")
		return
	}
	response := listLinksResponse{
		Links:  make([]linkItemResponse, 0, len(links)),
		Limit:  limit,
		Offset: offset,
		Total:  total,
	}
	for _, link := range links {
		response.Links = append(response.Links, linkItemResponse{
			ShortCode: link.ID,
			URL:       link.OriginalURL,
			Visits:    link.Visits,
			CreatedAt: link.CreatedAt.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *LinkHandler) getLink(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("short_code")
	link, err := h.service.Get(r.Context(), shortCode)
	if err != nil {
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	writeJSON(w, http.StatusOK, getLinkResponse{
		URL:    link.OriginalURL,
		Visits: link.Visits,
	})
}

func (h *LinkHandler) deleteLink(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("short_code")
	if err := h.service.Delete(r.Context(), shortCode); err != nil {
		if errors.Is(err, repository.ErrLinkNotFound) {
			writeError(w, http.StatusNotFound, "link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete link")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *LinkHandler) getStats(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("short_code")
	link, err := h.service.Stats(r.Context(), shortCode)
	if err != nil {
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	writeJSON(w, http.StatusOK, getStatsResponse{
		ShortCode: link.ID,
		URL:       link.OriginalURL,
		Visits:    link.Visits,
		CreatedAt: link.CreatedAt.Format(time.RFC3339),
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}
