package stun

import (
	"context"
	"fmt"
	"net"
	"time"

	pion "github.com/pion/stun/v3"
)

type StunAddressResolver struct {
	ctx        context.Context
	cancel     context.CancelFunc
	conn       *net.UDPConn
	serverAddr *net.UDPAddr
	publicAddr pion.XORMappedAddress
	msgCh      chan []byte
}

func NewStunAddressResolver(ctx context.Context, serverAddr string) (*StunAddressResolver, error) {
	addr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("resolve: %w", err)
	}

	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	swCtx, cancel := context.WithCancel(ctx)

	w := &StunAddressResolver{
		ctx:        swCtx,
		cancel:     cancel,
		conn:       conn,
		serverAddr: addr,
		msgCh:      make(chan []byte),
	}

	go w.listen()

	return w, nil
}

func (s *StunAddressResolver) Resolve() error {
	defer s.cancel()

	if err := s.sendBindRequest(); err != nil {
		s.conn.Close()
		return fmt.Errorf("send: %w", err)
	}

	if err := s.handle(); err != nil {
		return err
	}

	return nil
}

// Get addr
func (s *StunAddressResolver) PublicAddr() string {
	return s.publicAddr.String()
}

// Get ip
func (s *StunAddressResolver) PublicIP() string {
	return s.publicAddr.IP.String()
}

// Get port
func (s *StunAddressResolver) PublicPort() int {
	return s.publicAddr.Port
}

// Get conn
func (s *StunAddressResolver) Conn() *net.UDPConn {
	return s.conn
}

func (s *StunAddressResolver) handle() error {
	select {
	case <-s.ctx.Done():
		return nil

	case <-time.After(20 * time.Second):
		return fmt.Errorf("timeout")

	case message, ok := <-s.msgCh:
		if !ok {
			return nil
		}

		m := new(pion.Message)
		m.Raw = message

		if err := m.Decode(); err != nil {
			// log.Errorf("decode: %v", err)
			return err
		}

		var xorAddr pion.XORMappedAddress
		if err := xorAddr.GetFrom(m); err != nil {
			// log.Errorf("get: %v", err)
			return err
		}

		if s.publicAddr.String() != xorAddr.String() {
			s.publicAddr = xorAddr
			return nil
		}
	}

	return nil
}

// send bind request to stun server
func (s *StunAddressResolver) sendBindRequest() error {
	m := pion.MustBuild(pion.TransactionID, pion.BindingRequest)

	_, err := s.conn.WriteToUDP(m.Raw, s.serverAddr)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

// listen for response from stun server
func (s *StunAddressResolver) listen() {
	buf := make([]byte, 1024)
	go func() {
		defer close(s.msgCh)
		for {
			select {
			case <-s.ctx.Done():
				return

			default:
				n, _, err := s.conn.ReadFromUDP(buf)
				if err != nil {
					return
				}
				buf = buf[:n]

				s.msgCh <- buf
			}
		}
	}()
}
