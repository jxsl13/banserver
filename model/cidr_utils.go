package model

import (
	"net"
	"regexp"
)

var (
	cidrRegex = regexp.MustCompile(`^\s*([0-9.:a-fA-F\-\/]+)\s*.*$`)
)

// allow single ip addresses and CIDR ranges
func parseCIDR(line string) (_ *net.IPNet, ok bool) {
	matches := cidrRegex.FindStringSubmatch(line)
	if len(matches) == 0 {
		return nil, false
	}

	_, cidr, err := net.ParseCIDR(matches[1])
	if err == nil {
		return cidr, true
	}

	ip := net.ParseIP(matches[1])
	if ip == nil {
		return nil, false
	}

	var mask net.IPMask
	if ip.To4() != nil {
		// IPv4 /32 single ip msk
		mask = net.CIDRMask(32, 32)
	} else {
		// IPv6 - /128 single ip mask
		mask = net.CIDRMask(128, 128)
	}

	return &net.IPNet{
		IP:   ip,
		Mask: mask,
	}, true
}
