package model

import (
	"bufio"
	"context"
	"errors"
	"log"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/jxsl13/banserver/econ"
	"github.com/jxsl13/banserver/parser"
)

type Broker struct {
	mu        sync.RWMutex
	banserver *BanServer
	// filepath -> blocklist
	chatBlockRegex []*regexp.Regexp

	// default reasons and durations for bans
	permabanDuration time.Duration
	permabanReason   string
	chatBanDuration  time.Duration
	chatBanReason    string

	propagate bool

	serverMap map[string]*econ.Server

	// server -> all others
	others map[string][]string
}

func NewBroker(
	propagate bool,
	permaBanDuration time.Duration,
	permabanReason string, chatBanDuration time.Duration,
	chatBanReason string,
) *Broker {
	return &Broker{
		banserver:        NewBanServer(),
		serverMap:        make(map[string]*econ.Server),
		permabanDuration: permaBanDuration,
		permabanReason:   permabanReason,
		chatBanDuration:  chatBanDuration,
		chatBanReason:    chatBanReason,
		propagate:        propagate,
	}
}

func (p *Broker) Close() (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for name, s := range p.serverMap {
		err = errors.Join(err, s.Close())
		delete(p.serverMap, name)
	}

	return nil
}

func (p *Broker) AddBlacklistCIDRFile(file string) error {
	return p.banserver.AddBlacklistCIDRFile(file)
}

func (p *Broker) AddBlacklistChatFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	deduplicated := make(map[string]struct{})
	list := make([]*regexp.Regexp, 0)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		if _, ok := deduplicated[line]; ok {
			continue
		}
		deduplicated[line] = struct{}{}

		re, err := regexp.Compile(line)
		if err != nil {
			return err
		}
		list = append(list, re)
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.chatBlockRegex = append(p.chatBlockRegex, list...)

	log.Printf("added %d regular expressions from file %s", len(list), file)
	return nil
}

func (p *Broker) DialTo(ctx context.Context, addrPort, password string) error {
	log.Printf("connecting to server %s...", addrPort)
	server, err := econ.DialTo(ctx, addrPort, password, p.handle)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.serverMap[addrPort] = server
	p.setOthersMap()

	log.Printf("connected to server %s", addrPort)
	return nil
}

func (p *Broker) setOthersMap() {

	others := make(map[string][]string, len(p.serverMap))

	for addrPort := range p.serverMap {
		others[addrPort] = make([]string, 0, len(p.serverMap)-1)

		for other := range p.serverMap {
			if other == addrPort {
				continue
			}
			others[addrPort] = append(others[addrPort], other)
		}
	}

	p.others = others
}

func (p *Broker) othersOf(server string) []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return slices.Clone(p.others[server])
}

func (p *Broker) BanOnAll(triggeringServer string, playerIP string, duration time.Duration, reason string) (err error) {
	defer func() {
		if err != nil {
			log.Printf("error banning ip %s on all servers triggered by %s: %v", playerIP, triggeringServer, err)
		}
	}()

	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, s := range p.serverMap {
		others := p.othersOf(s.AddressPort())
		s.IgnoreBanPrapagation(playerIP, others...)

		err := s.BanIP(triggeringServer, playerIP, duration, reason)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Broker) UnbanOnAll(triggeringServer, playerIP string) (err error) {
	defer func() {
		if err != nil {
			log.Printf("error unbanning ip %s on all servers triggered by %s: %v", playerIP, triggeringServer, err)
		}
	}()

	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, s := range p.serverMap {
		others := p.othersOf(s.AddressPort())
		s.IgnoreBanPrapagation(playerIP, others...)

		err := s.UnbanIP(triggeringServer, playerIP)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Broker) BanOnOthers(triggeringServer, playerIP string, duration time.Duration, reason string) (err error) {
	defer func() {
		if err != nil {
			log.Printf("error banning ip %s on all other servers of %s: %v", playerIP, triggeringServer, err)
		}
	}()

	p.mu.RLock()
	defer p.mu.RUnlock()

	ts, ok := p.serverMap[triggeringServer]
	if !ok {
		panic("triggering server not found in server map: this is a programming error")
	}

	otherServers := p.othersOf(triggeringServer)
	ts.IgnoreBanPrapagation(playerIP, otherServers...)

	for _, other := range otherServers {
		s, ok := p.serverMap[other]
		if !ok {
			panic("other server not found in server map: this is a programming error")
		}

		// removes the ignore flag
		if s.IsIgnoredBanPropagation(triggeringServer, playerIP) {
			// remove flag from current server
			_ = ts.IsIgnoredBanPropagation(other, playerIP)
			continue
		}

		err := s.BanIP(triggeringServer, playerIP, duration, reason)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Broker) UnbanOnOthers(triggeringServer, playerIP string) (err error) {
	defer func() {
		if err != nil {
			log.Printf("error unbanning ip %s on all other servers of %s: %v", playerIP, triggeringServer, err)
		}
	}()

	p.mu.RLock()
	defer p.mu.RUnlock()

	ts, ok := p.serverMap[triggeringServer]
	if !ok {
		panic("triggering server not found in server map: this is a programming error")
	}

	otherServers := p.othersOf(triggeringServer)
	ts.IgnoreUnbanPrapagation(playerIP, otherServers...)

	for _, other := range otherServers {
		s, ok := p.serverMap[other]
		if !ok {
			panic("other server not found in server map: this is a programming error")
		}

		// removes the ignore flag
		if s.IsIgnoredUnbanPropagation(triggeringServer, playerIP) {
			// remove flag from current server
			_ = ts.IsIgnoredUnbanPropagation(other, playerIP)
			continue
		}

		err := s.UnbanIP(triggeringServer, playerIP)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Broker) handle(s *econ.Server, line string) {
	chat, ok := parser.ParseChatMessage(line)
	if ok {
		p.handleChat(s, chat)
		return
	}

	entered, ok := parser.ParseClientEntered(line)
	if ok {
		p.handleEntered(s, entered)
		return
	}

	dropped, ok := parser.ParseClientDropped(line)
	if ok {
		p.handleDropped(s, dropped)
		return
	}

	banned, ok := parser.ParseClientBanned(line)
	if ok {
		p.handleBanned(s, banned)
		return
	}

	unbanned, ok := parser.ParseClientUnbanned(line)
	if ok {
		p.handleUnbanned(s, unbanned)
		return
	}

	return
}

func (p *Broker) handleEntered(s *econ.Server, entered parser.ClientEntered) {
	banned, err := p.banserver.IsBanned(entered.IP)
	if err != nil {
		log.Printf("error checking if client is banned: %v", err)
		return
	}

	if banned {
		log.Printf("banned client %s entered server %s", entered.IP, s.AddressPort())
		// just ban on the server that the client tries to enter.
		// we do not want to propagate the ban to all other servers.
		// because we can just ban the IP once it tries to enter the other server.
		// this way we do not spam the ban list of all other servers.
		err := s.BanIP(s.AddressPort(), entered.IP, p.permabanDuration, p.permabanReason)
		if err != nil {
			log.Printf("error banning client %s on server %s: %v", entered.IP, s.AddressPort(), err)
			return
		}
	} else {
		log.Printf("client %s entered server %s", entered.IP, s.AddressPort())
	}
}

func (p *Broker) handleDropped(s *econ.Server, dropped parser.ClientDropped) {
	log.Printf("client %s dropped from server %s", dropped.IP, s.AddressPort())
}

func (p *Broker) handleBanned(s *econ.Server, banned parser.ClientBanned) {
	if !p.propagate {
		return
	}

	log.Printf("propagating client %s banned on server %s", banned.IP, s.AddressPort())

	// propagate ban to other servers
	err := p.BanOnOthers(s.AddressPort(), banned.IP, banned.Duration, banned.Reason)
	if err != nil {
		log.Printf("error propagating ban to other servers: %v", err)
	}
}

func (p *Broker) handleUnbanned(s *econ.Server, unbanned parser.ClientUnbanned) {
	if !p.propagate {
		return
	}

	log.Printf("propagating client %s unbanned on server %s", unbanned.IP, s.AddressPort())

	// propagate unban to other servers
	err := p.UnbanOnOthers(s.AddressPort(), unbanned.IP)
	if err != nil {
		log.Printf("error propagating unban to other servers: %v", err)
	}
}

func (p *Broker) handleChat(s *econ.Server, chat parser.ChatMessage) {
	for _, re := range p.chatBlockRegex {
		if !re.MatchString(chat.Message) {
			continue
		}

		log.Printf("banning client on all servers for chat message matching '%s': %s", re.String(), chat.Message)
		ip, ok := s.ClientIP(chat.ClientID)
		if !ok {
			log.Printf("error getting client ip for chat message: %v", chat.ClientID)
			return
		}

		if ip == "" {
			log.Printf("client ip is empty for chat message: %v", chat.ClientID)
			return
		}

		err := p.BanOnAll(s.AddressPort(), ip, p.chatBanDuration, p.chatBanReason)
		if err != nil {
			log.Printf("error banning client %s for chat message: %v", ip, err)
			return
		}
		return
	}
}
