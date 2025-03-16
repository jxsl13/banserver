package model

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/yl2chen/cidranger"
)

type BanServer struct {
	mu sync.RWMutex
	r  cidranger.Ranger
}

func NewBanServer() *BanServer {
	return &BanServer{
		r: cidranger.NewPCTrieRanger(),
	}
}

func (b *BanServer) HasIPs() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.r.Len() > 0
}

// AddBannedCIDR adds a new CIDR to the ban server, e.g. 192.168.1.0/24
func (b *BanServer) AddBannedCIDR(cidr string) error {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.r.Insert(cidranger.NewBasicRangerEntry(*network))
	return nil
}

// RemoveBannedCIDR removes a CIDR from the ban server
func (b *BanServer) RemoveBannedCIDR(cidr string) error {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.r.Remove(*network)
	return nil
}

// IsBanned checks if an IP is banned
func (b *BanServer) IsBanned(ip string) (banned bool, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to check if ip is banned: %w", err)
		}
	}()

	// teeworlds ipv6 addresses are enclosed in square brackets
	netIP := net.ParseIP(strings.Trim(ip, "[]"))
	if netIP == nil {
		return false, fmt.Errorf("invalid ip address: %s", ip)
	}

	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.r.Contains(netIP)
}

func (b *BanServer) AddBlacklistCIDRFile(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	cidrs := []*net.IPNet{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		cidr, ok := parseCIDR(line)
		if !ok {
			continue
		}

		cidrs = append(cidrs, cidr)
	}

	if len(cidrs) == 0 {
		return nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, cidr := range cidrs {
		b.r.Insert(cidranger.NewBasicRangerEntry(*cidr))
	}

	log.Printf("added %d CIDRs from file %s", len(cidrs), filePath)
	return nil
}
