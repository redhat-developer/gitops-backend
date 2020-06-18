package http

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// APIRouter is an HTTP API for accessing app configurations.
type APIRouter struct {
	*httprouter.Router
}

// Index returns the initial page.
func (a *APIRouter) Index(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"testing": "testing"})
}

// NewRouter creates and returns a new APIRouter.
func NewRouter() *APIRouter {
	api := &APIRouter{Router: httprouter.New()}
	api.HandlerFunc(http.MethodGet, "/", api.Index)
	return api
}
