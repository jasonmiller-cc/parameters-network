//go:build !linux

// Package service provides a no-op stub of NetworkService for non-Linux platforms.
// This allows the service to compile and unit-test on macOS/Windows without netlink.
package service

import "errors"

var errNotLinux = errors.New("network operations require Linux")

// StubNetworkService is a no-op implementation used on non-Linux platforms.
type StubNetworkService struct{}

// NewNetworkService returns the stub implementation on non-Linux platforms.
func NewNetworkService() NetworkService {
	return &StubNetworkService{}
}

func (s *StubNetworkService) ListInterfaces() ([]Interface, error)                      { return nil, errNotLinux }
func (s *StubNetworkService) GetInterface(name string) (*Interface, error)               { return nil, errNotLinux }
func (s *StubNetworkService) UpdateInterface(name string, u InterfaceUpdate) (*Interface, error) {
	return nil, errNotLinux
}
func (s *StubNetworkService) ListAddresses(iface string) ([]string, error)  { return nil, errNotLinux }
func (s *StubNetworkService) AddAddress(iface string, req AddAddressRequest) error {
	return errNotLinux
}
func (s *StubNetworkService) DeleteAddress(iface, address string) error     { return errNotLinux }
func (s *StubNetworkService) ListRoutes(q RouteQuery) ([]Route, error)      { return nil, errNotLinux }
func (s *StubNetworkService) AddRoute(req AddRouteRequest) error            { return errNotLinux }
func (s *StubNetworkService) DeleteRoute(req DeleteRouteRequest) error      { return errNotLinux }
func (s *StubNetworkService) ListNeighbors() ([]Neighbor, error)            { return nil, errNotLinux }
func (s *StubNetworkService) ListVLANs() ([]VLAN, error)                    { return nil, errNotLinux }
func (s *StubNetworkService) CreateVLAN(req CreateVLANRequest) (*VLAN, error) {
	return nil, errNotLinux
}
func (s *StubNetworkService) DeleteVLAN(name string) error                  { return errNotLinux }
