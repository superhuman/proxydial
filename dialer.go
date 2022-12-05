// Package proxydial provides a replacement for Go's builtin net.Dialer that
// is designed to be used in applications where you're proxying a user's request,
// or more generally making requests to URLs that you don't necessarily trust
// completely.
//
// Usually you will use the Dial function in an HTTP Client as follows:
//
//     client := http.Client{
//         Transport: &http.Transport{
//              Proxy: http.ProxyFromEnvironment,
//				Dial: proxydialer.Dial,
//              TLSHandshakeTimeout: 10 * time.Second
//         }
//     }
//
// The advantage of the proxydialer is that it prevents connections being made to
// internal IP addresses (e.g. 127.0.0.1, or 169.254.169.254) or to ports that could
// cause harm to the rest of the internet (e.g. 22, 25).
//
package proxydial

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

// Dialer lets you connect to external addresses. It's equivalent to net.Dialer
// from the go standard library, except that connections that are not to AllowedNets,
// not to AllowedPorts, or to BlockedRanges are aborted. It also does not yet support
// HappyEyeballs.
type Dialer struct {
	// AllowedNets is a whitelist of nets that connections may be made over. For http
	// this should be only []string{"tcp"}
	AllowedNets []string

	// AllowedPorts is a whitelist of ports that connections may be made to. For http
	// the usual suspects would be []int16{80, 443}, but some other ports are used
	// fairly frequently.
	AllowedPorts []int16

	// BlockedRanges is a black list of IP ranges that connections may not be made to
	// This should usually include net.ParseCIDR("127.0.0.1/8") to prevent connections
	// to localhost, in addition to any other subnets that you use internally.
	BlockedRanges []*net.IPNet

	// BlockPrivate blocks any IP on a private network
	BlockPrivate bool
	// BlockLinkLocal blocks any IP on a link local network
	BlockLinkLocal bool
	// BlockMulticast blocks any IP on a multicast network
	BlockMulticast bool
	// BlockUnspecified blocks any unspecified IP
	BlockUnspecified bool

	// Timeout is the maximum amount of time a dial will wait for
	// a connection to complete. If Deadline is also set, it may fail
	// earlier.
	//
	// The default is no timeout.
	//
	// When dialing a name with multiple IP addresses, the timeout
	// may be divided between them.
	//
	// With or without a timeout, the operating system may impose
	// its own earlier timeout. For instance, TCP timeouts are
	// often around 3 minutes.
	Timeout time.Duration

	// Deadline is the absolute point in time after which dials
	// will fail. If Timeout is set, it may fail earlier.
	// Zero means no deadline, or dependent on the operating system
	// as with the Timeout option.
	Deadline time.Time

	// KeepAlive specifies the keep-alive period for an active
	// network connection.
	// If zero, keep-alives are not enabled. Network protocols
	// that do not support keep-alives ignore this field.
	KeepAlive time.Duration

	// LocalAddr is the local address to use when dialing an
	// address. The address must be of a compatible type for the
	// network being dialed.
	// If nil, a local address is automatically chosen.
	LocalAddr net.Addr
}

func cidrRange(str string) *net.IPNet {

	_, net, err := net.ParseCIDR(str)
	if err != nil {
		panic(err)
	}

	return net
}

// DefaultDialer is a proxydialer designed for HTTP. It prevents connections to
// non-standard HTTP ports and internal IP addresses.
var DefaultDialer = Dialer{
	AllowedNets: []string{"tcp"},
	// Allow for the fairly common 8080 and 8443 ports too.
	AllowedPorts: []int16{80, 443, 8080, 8443},
	// From https://en.wikipedia.org/wiki/Reserved_IP_addresses
	BlockedRanges: []*net.IPNet{
		cidrRange("0.0.0.0/8"),
		cidrRange("10.0.0.0/8"),
		cidrRange("100.64.0.0/10"),
		cidrRange("127.0.0.0/8"),
		cidrRange("169.254.0.0/16"),
		cidrRange("172.16.0.0/12"),
		cidrRange("192.0.0.0/24"),
		cidrRange("192.0.2.0/24"),
		cidrRange("192.88.99.0/24"),
		cidrRange("192.168.0.0/16"),
		cidrRange("198.18.0.0/15"),
		cidrRange("198.51.100.0/24"),
		cidrRange("203.0.113.0/24"),
		cidrRange("224.0.0.0/4"),
		cidrRange("240.0.0.0/4"),
		cidrRange("255.255.255.255/32"),
		cidrRange("::/128"),
		cidrRange("::1/128"), // IPv6 loopback
		// cidrRange("::ffff:0:0/96"), IPv4 equivalents.
		cidrRange("100::/64"),
		cidrRange("64:ff9b::/96"),
		cidrRange("2001::/32"),
		cidrRange("2001:10::/28"),
		cidrRange("2001:20::/28"),
		cidrRange("2001:db8::/32"),
		cidrRange("2002::/16"),
		cidrRange("fc00::/7"),  // IPv6 unique local addr
		cidrRange("fe80::/10"), // IPv6 link-local
		cidrRange("ff00::/8"),  // IPv6 multicast
	},
	BlockMulticast:   true,
	BlockPrivate:     true,
	BlockLinkLocal:   true,
	BlockUnspecified: true,
}

// Dial creates a connection to the given address using DefaultDialer.Dial
func Dial(network, addr string) (net.Conn, error) {
	return DefaultDialer.Dial(network, addr)
}

func (d *Dialer) allowedNet(network string) bool {
	for _, net := range d.AllowedNets {
		if net == network {
			return true
		}
	}
	return false
}

func (d *Dialer) allowedPort(port int16) bool {
	for _, p := range d.AllowedPorts {
		if p == port {
			return true
		}
	}

	return false
}

func (d *Dialer) allowedIP(ip net.IP) bool {
	if d.BlockPrivate && (ip.IsPrivate() || ip.IsLoopback()) {
		return false
	}

	if d.BlockLinkLocal && (ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()) {
		return false
	}

	if d.BlockMulticast && (ip.IsMulticast() || ip.IsInterfaceLocalMulticast()) {
		return false
	}

	if d.BlockUnspecified && ip.IsUnspecified() {
		return false
	}

	for _, netrange := range d.BlockedRanges {
		if netrange.Contains(ip) {
			return false
		}
	}
	return true
}

func (d *Dialer) dialSerial(network, addr string, ips []net.IP, port string) (net.Conn, error) {

	dialer := net.Dialer{
		Timeout:   d.Timeout,
		Deadline:  d.Deadline,
		KeepAlive: d.KeepAlive,
		LocalAddr: d.LocalAddr,
	}

	// Ensure the deadline is set when a timeout is set, so that the total
	// amount of time used does not exceed the timeout even when multiple
	// requests are made.
	if dialer.Timeout > 0 {
		newDeadline := time.Now().Add(dialer.Timeout)
		if dialer.Deadline.IsZero() || newDeadline.Before(d.Deadline) {
			dialer.Deadline = newDeadline
		}
	}

	// Ensure that the timeout for each operation is small enough that
	// if connecting to the first address times out, the other addresses
	// will be tried.
	if !dialer.Deadline.IsZero() {
		totalTime := dialer.Deadline.Sub(time.Now())
		newTimeout := totalTime / time.Duration(len(ips))

		if newTimeout < 2*time.Second {
			newTimeout = 2 * time.Second
		}

		if dialer.Timeout == 0 || newTimeout < dialer.Timeout {
			dialer.Timeout = newTimeout
		}

	}

	var firstErr error
	for _, ip := range ips {
		conn, err := dialer.Dial(network, net.JoinHostPort(ip.String(), port))
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		return conn, nil
	}

	if firstErr == nil {
		firstErr = fmt.Errorf("dialer.Dial no IP addresses found: %s", addr)
	}

	return nil, firstErr
}

// Dial creates a connection to the given address. If the network is not in d.AllowedNetworks,
// or the port is not in d.AllowedPorts, or the IP address after DNS resolution is in b.BlockedRanges,
// then the connection will not be attempted.
func (d *Dialer) Dial(network, addr string) (net.Conn, error) {

	if !d.allowedNet(network) {
		return nil, fmt.Errorf("dialer.Dial %s %s: invalid net", network, addr)
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	portnum, err := parsePort(network, port)
	if err != nil {
		return nil, err
	}

	if !d.allowedPort(int16(portnum)) {
		return nil, fmt.Errorf("dialer.Dial %s: blocked port", addr)
	}

	ips, err := net.LookupIP(host)

	if err != nil {
		return nil, err
	}

	// Block any attempt to connect to any host that advertises an internal IP address.
	// TODO:CI — in the real world are there systems that advertise both their internal &
	// external IPs?
	for _, ip := range ips {
		if !d.allowedIP(ip) {
			return nil, fmt.Errorf("dialer.Dial %s: blocked range (%s)", addr, ip)
		}
	}

	return d.dialSerial(network, addr, ips, strconv.Itoa(portnum))

}
