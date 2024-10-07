package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-logr/logr"

	"github.com/naver/lobster/pkg/operator/server/controller"
)

const (
	PathSync = "/sync"
)

type InternalSyncHandler struct {
	Ctrl   controller.SinkController
	Logger logr.Logger
}

func (h InternalSyncHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func(r *http.Request) {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			h.Logger.Error(err, "failed to discard body")
		}
		if err := r.Body.Close(); err != nil {
			h.Logger.Error(err, "failed to close body")
		}
	}(r)

	switch r.Method {
	case http.MethodGet:
		h.handleGet(w)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h InternalSyncHandler) handleGet(w http.ResponseWriter) {
	sinks, err := h.Ctrl.List("", "", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(sinks) == 0 {
		http.Error(w, "no sinks", http.StatusNoContent)
		return
	}

	data, err := json.Marshal(sinks)
	if err != nil {
		handleError(w, err)
		return
	}

	if _, err := w.Write(data); err != nil {
		h.Logger.Error(err, "failed to write contents")
	}
}
