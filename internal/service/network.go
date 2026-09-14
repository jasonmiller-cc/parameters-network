// Package service defines the NetworkService interface and shared data types
// for network interface and routing management.
package service

// Interface represents a network interface with statistics.
type Interface struct {
	Name      string   `json:"name"`
	MAC       string   `json:"mac"`
	MTU       int      `json:"mtu"`
	Up        bool     `json:"up"`
	Addresses []string `json:"addresses"`
	RxBytes   int64    `json:"rx_bytes"`
	TxBytes   int64    `json:"tx_bytes"`
}

// Route represents a routing table entry.
type Route struct {
	Dst      string `json:"dst"`
	Gateway  string `json:"gateway"`
	Dev      string `json:"dev"`
	Metric   int    `json:"metric"`
	Table    int    `json:"table"`
	Protocol string `json:"protocol"`
}

// Neighbor represents an ARP or NDP neighbor table entry.
type Neighbor struct {
	IP    string `json:"ip"`
	MAC   string `json:"mac"`
	Dev   string `json:"dev"`
	State string `json:"state"`
}

// VLAN represents a VLAN sub-interface.
type VLAN struct {
	Name      string   `json:"name"`
	Parent    string   `json:"parent"`
	VLANID    int      `json:"vlan_id"`
	Addresses []string `json:"addresses"`
}

// InterfaceUpdate carries fields that may be changed on an existing interface.
type InterfaceUpdate struct {
	MTU *int  `json:"mtu,omitempty"`
	Up  *bool `json:"up,omitempty"`
}

// AddRouteRequest carries parameters for a new route.
type AddRouteRequest struct {
	Dst     string `json:"dst"`
	Gateway string `json:"gateway"`
	Dev     string `json:"dev"`
	Metric  int    `json:"metric"`
	Table   int    `json:"table"`
}

// DeleteRouteRequest carries parameters to identify a route for deletion.
type DeleteRouteRequest = AddRouteRequest

// AddAddressRequest carries an address to add to an interface.
type AddAddressRequest struct {
	Address   string `json:"address"`
	PrefixLen int    `json:"prefix_len"`
}

// CreateVLANRequest carries parameters for a new VLAN interface.
type CreateVLANRequest struct {
	Parent string `json:"parent"`
	VLANID int    `json:"vlan_id"`
	Name   string `json:"name"`
}

// RouteQuery holds optional filters for listing routes.
type RouteQuery struct {
	Table  int // 0 = main table
	Family int // 0 = all, syscall.AF_INET, syscall.AF_INET6
}

// NetworkService is the interface that abstracts Linux netlink operations.
// A real Linux implementation lives in network_linux.go; a cross-platform stub
// lives in network_stub.go.
type NetworkService interface {
	// Interfaces
	ListInterfaces() ([]Interface, error)
	GetInterface(name string) (*Interface, error)
	UpdateInterface(name string, update InterfaceUpdate) (*Interface, error)

	// Addresses
	ListAddresses(iface string) ([]string, error)
	AddAddress(iface string, req AddAddressRequest) error
	DeleteAddress(iface, address string) error

	// Routes
	ListRoutes(q RouteQuery) ([]Route, error)
	AddRoute(req AddRouteRequest) error
	DeleteRoute(req DeleteRouteRequest) error

	// Neighbors
	ListNeighbors() ([]Neighbor, error)

	// VLANs
	ListVLANs() ([]VLAN, error)
	CreateVLAN(req CreateVLANRequest) (*VLAN, error)
	DeleteVLAN(name string) error
}
