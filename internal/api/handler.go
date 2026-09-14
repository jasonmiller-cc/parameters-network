// Package api wires HTTP routes for the parameters-network REST API.
package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/jasonmiller-cc/parameters-core/pkg/response"
	"github.com/jasonmiller-cc/parameters-network/internal/service"
)

// Handler holds the dependencies for all network API endpoints.
type Handler struct {
	svc service.NetworkService
}

// New creates a Handler backed by the given NetworkService.
func New(svc service.NetworkService) *Handler {
	return &Handler{svc: svc}
}

// Register binds all routes to mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/interfaces", h.listInterfaces)
	mux.HandleFunc("GET /api/v1/interfaces/{name}", h.getInterface)
	mux.HandleFunc("PUT /api/v1/interfaces/{name}", h.updateInterface)
	mux.HandleFunc("GET /api/v1/interfaces/{name}/addresses", h.listAddresses)
	mux.HandleFunc("POST /api/v1/interfaces/{name}/addresses", h.addAddress)
	mux.HandleFunc("DELETE /api/v1/interfaces/{name}/addresses/{address}", h.deleteAddress)

	mux.HandleFunc("GET /api/v1/routes", h.listRoutes)
	mux.HandleFunc("POST /api/v1/routes", h.addRoute)
	mux.HandleFunc("DELETE /api/v1/routes", h.deleteRoute)

	mux.HandleFunc("GET /api/v1/neighbors", h.listNeighbors)

	mux.HandleFunc("GET /api/v1/vlans", h.listVLANs)
	mux.HandleFunc("POST /api/v1/vlans", h.createVLAN)
	mux.HandleFunc("DELETE /api/v1/vlans/{name}", h.deleteVLAN)

	mux.HandleFunc("GET /api/v1/stats", h.getStats)
}

// --- Interfaces ---

func (h *Handler) listInterfaces(w http.ResponseWriter, r *http.Request) {
	ifaces, err := h.svc.ListInterfaces()
	if err != nil {
		response.Err(w, err)
		return
	}
	response.OK(w, ifaces)
}

func (h *Handler) getInterface(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	iface, err := h.svc.GetInterface(name)
	if err != nil {
		response.Err(w, err)
		return
	}
	response.OK(w, iface)
}

func (h *Handler) updateInterface(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var update service.InterfaceUpdate
	if err := response.DecodeJSON(r, &update); err != nil {
		response.Err(w, err)
		return
	}
	iface, err := h.svc.UpdateInterface(name, update)
	if err != nil {
		response.Err(w, err)
		return
	}
	response.OK(w, iface)
}

// --- Addresses ---

func (h *Handler) listAddresses(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	addrs, err := h.svc.ListAddresses(name)
	if err != nil {
		response.Err(w, err)
		return
	}
	response.OK(w, addrs)
}

func (h *Handler) addAddress(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var req service.AddAddressRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Err(w, err)
		return
	}
	if err := h.svc.AddAddress(name, req); err != nil {
		response.Err(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) deleteAddress(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	// The address may contain slashes (IPv6) so we join everything after the prefix.
	address := r.PathValue("address")
	if err := h.svc.DeleteAddress(name, address); err != nil {
		response.Err(w, err)
		return
	}
	response.NoContent(w)
}

// --- Routes ---

func (h *Handler) listRoutes(w http.ResponseWriter, r *http.Request) {
	q := service.RouteQuery{}
	if t := r.URL.Query().Get("table"); t != "" {
		if v, err := strconv.Atoi(t); err == nil {
			q.Table = v
		}
	}
	if f := r.URL.Query().Get("family"); f != "" {
		switch strings.ToLower(f) {
		case "4", "inet", "ipv4":
			q.Family = 2 // syscall.AF_INET
		case "6", "inet6", "ipv6":
			q.Family = 10 // syscall.AF_INET6
		}
	}
	routes, err := h.svc.ListRoutes(q)
	if err != nil {
		response.Err(w, err)
		return
	}
	response.OK(w, routes)
}

func (h *Handler) addRoute(w http.ResponseWriter, r *http.Request) {
	var req service.AddRouteRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Err(w, err)
		return
	}
	if err := h.svc.AddRoute(req); err != nil {
		response.Err(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) deleteRoute(w http.ResponseWriter, r *http.Request) {
	var req service.DeleteRouteRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Err(w, err)
		return
	}
	if err := h.svc.DeleteRoute(req); err != nil {
		response.Err(w, err)
		return
	}
	response.NoContent(w)
}

// --- Neighbors ---

func (h *Handler) listNeighbors(w http.ResponseWriter, r *http.Request) {
	neighbors, err := h.svc.ListNeighbors()
	if err != nil {
		response.Err(w, err)
		return
	}
	response.OK(w, neighbors)
}

// --- VLANs ---

func (h *Handler) listVLANs(w http.ResponseWriter, r *http.Request) {
	vlans, err := h.svc.ListVLANs()
	if err != nil {
		response.Err(w, err)
		return
	}
	response.OK(w, vlans)
}

func (h *Handler) createVLAN(w http.ResponseWriter, r *http.Request) {
	var req service.CreateVLANRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Err(w, err)
		return
	}
	vlan, err := h.svc.CreateVLAN(req)
	if err != nil {
		response.Err(w, err)
		return
	}
	response.Created(w, vlan)
}

func (h *Handler) deleteVLAN(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := h.svc.DeleteVLAN(name); err != nil {
		response.Err(w, err)
		return
	}
	response.NoContent(w)
}

// --- Stats ---

// statsEntry is a per-interface rx/tx summary returned by GET /api/v1/stats.
type statsEntry struct {
	Name    string `json:"name"`
	RxBytes int64  `json:"rx_bytes"`
	TxBytes int64  `json:"tx_bytes"`
}

func (h *Handler) getStats(w http.ResponseWriter, r *http.Request) {
	ifaces, err := h.svc.ListInterfaces()
	if err != nil {
		response.Err(w, err)
		return
	}
	stats := make([]statsEntry, 0, len(ifaces))
	for _, iface := range ifaces {
		stats = append(stats, statsEntry{
			Name:    iface.Name,
			RxBytes: iface.RxBytes,
			TxBytes: iface.TxBytes,
		})
	}
	response.OK(w, stats)
}
