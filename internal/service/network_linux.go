//go:build linux

package service

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// LinuxNetworkService implements NetworkService using the Linux netlink API.
type LinuxNetworkService struct{}

// NewNetworkService returns the Linux netlink-backed implementation.
func NewNetworkService() NetworkService {
	return &LinuxNetworkService{}
}

// --- Interfaces ---

func (s *LinuxNetworkService) ListInterfaces() ([]Interface, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("netlink link list: %w", err)
	}
	out := make([]Interface, 0, len(links))
	for _, l := range links {
		iface, err := s.linkToInterface(l)
		if err != nil {
			continue
		}
		out = append(out, *iface)
	}
	return out, nil
}

func (s *LinuxNetworkService) GetInterface(name string) (*Interface, error) {
	l, err := netlink.LinkByName(name)
	if err != nil {
		return nil, fmt.Errorf("netlink link by name %q: %w", name, err)
	}
	return s.linkToInterface(l)
}

func (s *LinuxNetworkService) UpdateInterface(name string, update InterfaceUpdate) (*Interface, error) {
	l, err := netlink.LinkByName(name)
	if err != nil {
		return nil, fmt.Errorf("netlink link by name %q: %w", name, err)
	}
	if update.MTU != nil {
		if err := netlink.LinkSetMTU(l, *update.MTU); err != nil {
			return nil, fmt.Errorf("set mtu: %w", err)
		}
	}
	if update.Up != nil {
		if *update.Up {
			if err := netlink.LinkSetUp(l); err != nil {
				return nil, fmt.Errorf("link up: %w", err)
			}
		} else {
			if err := netlink.LinkSetDown(l); err != nil {
				return nil, fmt.Errorf("link down: %w", err)
			}
		}
	}
	return s.GetInterface(name)
}

func (s *LinuxNetworkService) linkToInterface(l netlink.Link) (*Interface, error) {
	attrs := l.Attrs()
	addrs, err := netlink.AddrList(l, netlink.FAMILY_ALL)
	if err != nil {
		addrs = nil
	}
	addrStrs := make([]string, 0, len(addrs))
	for _, a := range addrs {
		addrStrs = append(addrStrs, a.IPNet.String())
	}

	var rxBytes, txBytes int64
	if attrs.Statistics != nil {
		rxBytes = int64(attrs.Statistics.RxBytes)
		txBytes = int64(attrs.Statistics.TxBytes)
	}

	mac := ""
	if hw := attrs.HardwareAddr; hw != nil {
		mac = hw.String()
	}

	return &Interface{
		Name:      attrs.Name,
		MAC:       mac,
		MTU:       attrs.MTU,
		Up:        attrs.Flags&net.FlagUp != 0,
		Addresses: addrStrs,
		RxBytes:   rxBytes,
		TxBytes:   txBytes,
	}, nil
}

// --- Addresses ---

func (s *LinuxNetworkService) ListAddresses(iface string) ([]string, error) {
	l, err := netlink.LinkByName(iface)
	if err != nil {
		return nil, fmt.Errorf("link by name %q: %w", iface, err)
	}
	addrs, err := netlink.AddrList(l, netlink.FAMILY_ALL)
	if err != nil {
		return nil, fmt.Errorf("addr list: %w", err)
	}
	out := make([]string, 0, len(addrs))
	for _, a := range addrs {
		out = append(out, a.IPNet.String())
	}
	return out, nil
}

func (s *LinuxNetworkService) AddAddress(iface string, req AddAddressRequest) error {
	l, err := netlink.LinkByName(iface)
	if err != nil {
		return fmt.Errorf("link by name %q: %w", iface, err)
	}
	ip := net.ParseIP(req.Address)
	if ip == nil {
		return fmt.Errorf("invalid IP address %q", req.Address)
	}
	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(req.PrefixLen, 32),
		},
	}
	if ip.To4() == nil {
		addr.IPNet.Mask = net.CIDRMask(req.PrefixLen, 128)
	}
	return netlink.AddrAdd(l, addr)
}

func (s *LinuxNetworkService) DeleteAddress(iface, address string) error {
	l, err := netlink.LinkByName(iface)
	if err != nil {
		return fmt.Errorf("link by name %q: %w", iface, err)
	}
	addrs, err := netlink.AddrList(l, netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("addr list: %w", err)
	}
	for _, a := range addrs {
		if a.IPNet.IP.String() == address || a.IPNet.String() == address {
			return netlink.AddrDel(l, &a)
		}
	}
	return fmt.Errorf("address %q not found on %q", address, iface)
}

// --- Routes ---

func (s *LinuxNetworkService) ListRoutes(q RouteQuery) ([]Route, error) {
	family := q.Family
	if family == 0 {
		family = netlink.FAMILY_ALL
	}
	filter := &netlink.Route{}
	var filterMask uint64
	if q.Table != 0 {
		filter.Table = q.Table
		filterMask |= netlink.RT_FILTER_TABLE
	}
	routes, err := netlink.RouteListFiltered(family, filter, filterMask)
	if err != nil {
		return nil, fmt.Errorf("route list: %w", err)
	}
	out := make([]Route, 0, len(routes))
	for _, r := range routes {
		out = append(out, routeToRoute(r))
	}
	return out, nil
}

func routeToRoute(r netlink.Route) Route {
	dst := ""
	if r.Dst != nil {
		dst = r.Dst.String()
	}
	gw := ""
	if r.Gw != nil {
		gw = r.Gw.String()
	}
	dev := ""
	if l, err := netlink.LinkByIndex(r.LinkIndex); err == nil {
		dev = l.Attrs().Name
	}
	return Route{
		Dst:      dst,
		Gateway:  gw,
		Dev:      dev,
		Metric:   r.Priority,
		Table:    r.Table,
		Protocol: fmt.Sprintf("%d", r.Protocol),
	}
}

func (s *LinuxNetworkService) AddRoute(req AddRouteRequest) error {
	_, dst, err := net.ParseCIDR(req.Dst)
	if err != nil {
		return fmt.Errorf("parse dst %q: %w", req.Dst, err)
	}
	route := &netlink.Route{
		Dst:      dst,
		Priority: req.Metric,
		Table:    req.Table,
	}
	if req.Gateway != "" {
		route.Gw = net.ParseIP(req.Gateway)
	}
	if req.Dev != "" {
		l, err := netlink.LinkByName(req.Dev)
		if err != nil {
			return fmt.Errorf("link by name %q: %w", req.Dev, err)
		}
		route.LinkIndex = l.Attrs().Index
	}
	return netlink.RouteAdd(route)
}

func (s *LinuxNetworkService) DeleteRoute(req DeleteRouteRequest) error {
	_, dst, err := net.ParseCIDR(req.Dst)
	if err != nil {
		return fmt.Errorf("parse dst %q: %w", req.Dst, err)
	}
	route := &netlink.Route{
		Dst:      dst,
		Priority: req.Metric,
		Table:    req.Table,
	}
	if req.Gateway != "" {
		route.Gw = net.ParseIP(req.Gateway)
	}
	if req.Dev != "" {
		l, err := netlink.LinkByName(req.Dev)
		if err != nil {
			return fmt.Errorf("link by name %q: %w", req.Dev, err)
		}
		route.LinkIndex = l.Attrs().Index
	}
	return netlink.RouteDel(route)
}

// --- Neighbors ---

func (s *LinuxNetworkService) ListNeighbors() ([]Neighbor, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("link list: %w", err)
	}
	var out []Neighbor
	for _, l := range links {
		neighbors, err := netlink.NeighList(l.Attrs().Index, netlink.FAMILY_ALL)
		if err != nil {
			continue
		}
		for _, n := range neighbors {
			mac := ""
			if n.HardwareAddr != nil {
				mac = n.HardwareAddr.String()
			}
			out = append(out, Neighbor{
				IP:    n.IP.String(),
				MAC:   mac,
				Dev:   l.Attrs().Name,
				State: neighStateToString(n.State),
			})
		}
	}
	return out, nil
}

func neighStateToString(state int) string {
	switch state {
	case unix.NUD_REACHABLE:
		return "reachable"
	case unix.NUD_STALE:
		return "stale"
	case unix.NUD_DELAY:
		return "delay"
	case unix.NUD_PROBE:
		return "probe"
	case unix.NUD_FAILED:
		return "failed"
	case unix.NUD_NOARP:
		return "noarp"
	case unix.NUD_PERMANENT:
		return "permanent"
	default:
		return fmt.Sprintf("unknown(%d)", state)
	}
}

// --- VLANs ---

func (s *LinuxNetworkService) ListVLANs() ([]VLAN, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("link list: %w", err)
	}
	var out []VLAN
	for _, l := range links {
		vlan, ok := l.(*netlink.Vlan)
		if !ok {
			continue
		}
		parent := ""
		if pl, err := netlink.LinkByIndex(vlan.ParentIndex); err == nil {
			parent = pl.Attrs().Name
		}
		addrs, _ := netlink.AddrList(l, netlink.FAMILY_ALL)
		addrStrs := make([]string, 0, len(addrs))
		for _, a := range addrs {
			addrStrs = append(addrStrs, a.IPNet.String())
		}
		out = append(out, VLAN{
			Name:      vlan.Attrs().Name,
			Parent:    parent,
			VLANID:    vlan.VlanId,
			Addresses: addrStrs,
		})
	}
	return out, nil
}

func (s *LinuxNetworkService) CreateVLAN(req CreateVLANRequest) (*VLAN, error) {
	parent, err := netlink.LinkByName(req.Parent)
	if err != nil {
		return nil, fmt.Errorf("parent link %q: %w", req.Parent, err)
	}
	name := req.Name
	if name == "" {
		name = fmt.Sprintf("%s.%d", req.Parent, req.VLANID)
	}
	la := netlink.NewLinkAttrs()
	la.Name = name
	la.ParentIndex = parent.Attrs().Index
	vlan := &netlink.Vlan{
		LinkAttrs: la,
		VlanId:    req.VLANID,
	}
	if err := netlink.LinkAdd(vlan); err != nil {
		return nil, fmt.Errorf("link add vlan: %w", err)
	}
	return &VLAN{
		Name:      name,
		Parent:    req.Parent,
		VLANID:    req.VLANID,
		Addresses: []string{},
	}, nil
}

func (s *LinuxNetworkService) DeleteVLAN(name string) error {
	l, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("link by name %q: %w", name, err)
	}
	if _, ok := l.(*netlink.Vlan); !ok {
		return fmt.Errorf("%q is not a VLAN interface", name)
	}
	return netlink.LinkDel(l)
}
