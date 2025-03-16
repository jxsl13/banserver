package econ

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jxsl13/banserver/parser"
	"github.com/teeworlds-go/econ"
)

func DialTo(ctx context.Context, addrPort, password string, handler LineHandler) (_ *Server, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	conn, err := econ.DialTo(addrPort, password, econ.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	s := &Server{
		ctx:         ctx,
		cancel:      cancel,
		addrPort:    addrPort,
		password:    password,
		conn:        conn,
		lineChan:    make(chan string),
		commandChan: make(chan string),
		clients:     make(map[int]string),

		ignoredBanPropagation:   make(map[string]map[string]struct{}),
		ignoredUnbanPropagation: make(map[string]map[string]struct{}),
	}

	s.wg.Add(3) // 3 goroutines are started
	go s.asyncReadLine()
	go s.asyncWriteLine()
	go s.asyncProcess(handler)

	return s, nil
}

type Server struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	addrPort string
	password string

	conn *econ.Conn

	lineChan    chan string
	commandChan chan string

	// ID -> IP
	mu      sync.Mutex
	clients map[int]string

	// server -> ip
	ignoredBanPropagation   map[string]map[string]struct{}
	ignoredUnbanPropagation map[string]map[string]struct{}
}

func (s *Server) Close() error {
	s.cancel()
	err := s.conn.Close()
	s.wg.Wait()
	return err
}

func (s *Server) AddressPort() string {
	return s.addrPort
}

func (s *Server) ClientIP(id int) (ip string, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ip, ok = s.clients[id]
	return ip, ok
}

func (s *Server) send(command string) error {

	select {
	case <-s.ctx.Done():
		return fmt.Errorf("failed to send commands %q to %s: %v", command, s.addrPort, s.ctx.Err())
	case s.commandChan <- command:
		return nil
	}
}

func (s *Server) IgnoreBanPrapagation(bannedIP string, sourceServers ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, server := range sourceServers {

		if _, ok := s.ignoredBanPropagation[server]; !ok {
			s.ignoredBanPropagation[server] = make(map[string]struct{})
		}
		s.ignoredBanPropagation[server][bannedIP] = struct{}{}
	}
}

func (s *Server) IsIgnoredBanPropagation(triggeringServer, bannedIP string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	ignoredIPs, ok := s.ignoredBanPropagation[triggeringServer]
	if !ok {
		return false
	}

	if ignoredIPs == nil {
		return false
	}

	_, ok = ignoredIPs[bannedIP]
	if ok {
		delete(ignoredIPs, bannedIP)
	}
	return ok
}

func (s *Server) IgnoreUnbanPrapagation(unbannedIP string, sourceServers ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, server := range sourceServers {
		if _, ok := s.ignoredUnbanPropagation[server]; !ok {
			s.ignoredUnbanPropagation[server] = make(map[string]struct{})
		}

		s.ignoredUnbanPropagation[server][unbannedIP] = struct{}{}
	}
}

func (s *Server) IsIgnoredUnbanPropagation(triggeringServer, unbannedIP string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	ignoredIPs, ok := s.ignoredUnbanPropagation[triggeringServer]
	if !ok {
		return false
	}

	if ignoredIPs == nil {
		return false
	}

	_, ok = ignoredIPs[unbannedIP]
	if ok {
		delete(ignoredIPs, unbannedIP)
	}
	return ok
}

func (s *Server) BanIP(triggeringServer string, playerIP string, duration time.Duration, reason string) error {
	if playerIP == "" {
		return fmt.Errorf("ban failed on server %s: empty player ip", s.addrPort)
	}

	return s.send(fmt.Sprintf("ban %s %d %s", playerIP, int(duration.Minutes()), reason))
}

func (s *Server) UnbanIP(triggeringServer string, playerIP string) error {
	if playerIP == "" {
		return fmt.Errorf("unban failed on server %s: empty player ip", s.addrPort)
	}

	return s.send(fmt.Sprintf("unban %s", playerIP))
}

func (s *Server) asyncReadLine() {
	defer func() {
		close(s.lineChan)
		s.wg.Done()
		log.Println("line reader closed")
	}()

	var (
		line string
		err  error
	)
	for {
		select {
		case <-s.ctx.Done():
			log.Printf("closing line reader: %v", s.ctx.Err())
			return
		default:
			line, err = s.conn.ReadLine()
			if err != nil {
				if errors.Is(err, context.Canceled) {
					log.Printf("closing line reader: %v", err)
					return
				}
				log.Printf("failed to read line: %v", err)
			}
			s.lineChan <- line
		}
	}
}

func (s *Server) asyncWriteLine() {
	defer func() {
		s.wg.Done()
		log.Printf("command writer of %s closed", s.addrPort)
	}()

	var err error
	for {
		select {
		case <-s.ctx.Done():
			log.Printf("closing command writer of %s: %v", s.addrPort, s.ctx.Err())
			return
		case command, ok := <-s.commandChan:
			if !ok {
				log.Printf("command channel of %s closed", s.addrPort)
				return
			}
			err = s.conn.WriteLine(command)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					log.Printf("closing command writer of %s: %v", s.addrPort, s.ctx.Err())
					return
				}
				log.Printf("failed to write line to %s: %v", s.addrPort, err)
			}
		}
	}
}

type LineHandler func(server *Server, line string)

func (s *Server) asyncProcess(process func(s *Server, line string)) {
	defer func() {
		log.Printf("closing line processor of %s", s.addrPort)
		s.wg.Done()
		log.Printf("line processor of %s closed", s.addrPort)
	}()

	for {
		line, ok := tryRead(s.ctx, s.lineChan)
		if !ok {
			break
		}

		if entered, ok := parser.ParseClientEntered(line); ok {
			s.mu.Lock()
			s.clients[entered.ClientID] = entered.IP
			s.mu.Unlock()

			// allow the handler to process the line as well
		} else if dropped, ok := parser.ParseClientDropped(line); ok {
			s.mu.Lock()
			delete(s.clients, dropped.ClientID)
			s.mu.Unlock()
			// allow the handler to process the line as well
		}

		process(s, line)
	}

}

func tryRead(ctx context.Context, lineChan <-chan string) (line string, ok bool) {
	select {
	case <-ctx.Done():
		return "", false
	case line, ok = <-lineChan:
		return line, ok
	}
}
