package proxydial

// Everything in this file was copy-pasted from go/src/net (v1.5)
import "net"

// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// parsePort parses port as a network service port number for both
// TCP and UDP.
func parsePort(network, port string) (int, error) {
	p, i, ok := dtoi(port, 0)
	if !ok || i != len(port) {
		var err error
		p, err = net.LookupPort(network, port)
		if err != nil {
			return 0, err
		}
	}
	if p < 0 || p > 0xFFFF {
		return 0, &net.AddrError{Err: "invalid port", Addr: port}
	}
	return p, nil
}

const big = 0xFFFFFF

// Decimal to integer starting at &s[i0].
// Returns number, new offset, success.
func dtoi(s string, i0 int) (n int, i int, ok bool) {
	n = 0
	for i = i0; i < len(s) && '0' <= s[i] && s[i] <= '9'; i++ {
		n = n*10 + int(s[i]-'0')
		if n >= big {
			return 0, i, false
		}
	}
	if i == i0 {
		return 0, i, false
	}
	return n, i, true
}
