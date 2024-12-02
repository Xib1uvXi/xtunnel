package stun

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	pion "github.com/pion/stun/v3"
)

// implements the STUN RFC 5780
// https://tools.ietf.org/html/rfc5780
// from github.com/pion/stun

const (
	// the number of seconds to wait for STUN server's response
	timeout = 10
)

var (
	errResponseMessage = errors.New("error reading from response message channel")
	errTimedOut        = errors.New("timed out waiting for response")
	errNoOtherAddress  = errors.New("no OTHER-ADDRESS in message")
)

// RFC5780: 4.3.  Determining NAT Mapping Behavior
func MappingTests(ctx context.Context, addrStr string) (NatBehavior, error) {
	mapTestConn, err := connect(ctx, addrStr)
	if err != nil {
		return NatBehavior_NAT_BEHAVIOR_UNKNOWN, err
	}

	defer mapTestConn.Close()

	// Test I: Regular binding request
	request := pion.MustBuild(pion.TransactionID, pion.BindingRequest)

	resp, err := mapTestConn.roundTrip(request, mapTestConn.RemoteAddr)
	if err != nil {
		return NatBehavior_NAT_BEHAVIOR_UNKNOWN, err
	}

	// Parse response message for XOR-MAPPED-ADDRESS and make sure OTHER-ADDRESS valid
	resps1 := parse(resp)
	if resps1.xorAddr == nil || resps1.otherAddr == nil {
		return NatBehavior_NAT_BEHAVIOR_UNKNOWN, errNoOtherAddress
	}
	addr, err := net.ResolveUDPAddr("udp4", resps1.otherAddr.String())
	if err != nil {
		return NatBehavior_NAT_BEHAVIOR_UNKNOWN, err
	}
	mapTestConn.OtherAddr = addr
	//// log.Infof("Received XOR-MAPPED-ADDRESS: %v", resps1.xorAddr)

	// Assert mapping behavior
	if resps1.xorAddr.String() == mapTestConn.LocalAddr.String() {
		//// log.Warn("=> NAT mapping behavior: endpoint independent (no NAT)")
		return NatBehavior_NAT_BEHAVIOR_ENDPOINT_INDEPENDENT, nil
	}

	// Test II: Send binding request to the other address but primary port
	oaddr := *mapTestConn.OtherAddr
	oaddr.Port = mapTestConn.RemoteAddr.Port
	resp, err = mapTestConn.roundTrip(request, &oaddr)
	if err != nil {
		return NatBehavior_NAT_BEHAVIOR_UNKNOWN, err
	}

	// Assert mapping behavior
	resps2 := parse(resp)
	if resps2.xorAddr.String() == resps1.xorAddr.String() {
		//// log.Warn("=> NAT mapping behavior: endpoint independent")
		return NatBehavior_NAT_BEHAVIOR_ENDPOINT_INDEPENDENT, nil
	}

	// Test III: Send binding request to the other address and port
	resp, err = mapTestConn.roundTrip(request, mapTestConn.OtherAddr)
	if err != nil {
		return NatBehavior_NAT_BEHAVIOR_UNKNOWN, err
	}

	// Assert mapping behavior
	resps3 := parse(resp)
	if resps3.xorAddr.String() == resps2.xorAddr.String() {
		//// log.Warn("=> NAT mapping behavior: address dependent")
		return NatBehavior_NAT_BEHAVIOR_ADDRESS_DEPENDENT, nil
	} else {
		//// log.Warn("=> NAT mapping behavior: address and port dependent")
		return NatBehavior_NAT_BEHAVIOR_ADDRESS_AND_PORT_DEPENDENT, nil
	}
}

// RFC5780: 4.4.  Determining NAT Filtering Behavior
func FilteringTests(ctx context.Context, addrStr string) (NatBehavior, error) {
	mapTestConn, err := connect(ctx, addrStr)
	if err != nil {
		// log.Warnf("Error creating STUN connection: %s", err)
		return NatBehavior_NAT_BEHAVIOR_UNKNOWN, err
	}

	defer mapTestConn.Close()

	// Test I: Regular binding request
	request := pion.MustBuild(pion.TransactionID, pion.BindingRequest)

	resp, err := mapTestConn.roundTrip(request, mapTestConn.RemoteAddr)
	if err != nil || errors.Is(err, errTimedOut) {
		return NatBehavior_NAT_BEHAVIOR_UNKNOWN, err
	}
	resps := parse(resp)
	if resps.xorAddr == nil || resps.otherAddr == nil {
		return NatBehavior_NAT_BEHAVIOR_UNKNOWN, errNoOtherAddress
	}
	addr, err := net.ResolveUDPAddr("udp4", resps.otherAddr.String())
	if err != nil {
		return NatBehavior_NAT_BEHAVIOR_UNKNOWN, err
	}
	mapTestConn.OtherAddr = addr

	// Test II: Request to change both IP and port
	request = pion.MustBuild(pion.TransactionID, pion.BindingRequest)
	request.Add(pion.AttrChangeRequest, []byte{0x00, 0x00, 0x00, 0x06})

	resp, err = mapTestConn.roundTrip(request, mapTestConn.RemoteAddr)
	if err == nil {
		parse(resp) // just to print out the resp
		return NatBehavior_NAT_BEHAVIOR_ENDPOINT_INDEPENDENT, nil
	} else if !errors.Is(err, errTimedOut) {
		return NatBehavior_NAT_BEHAVIOR_UNKNOWN, err // something else went wrong
	}

	// Test III: Request to change port only
	request = pion.MustBuild(pion.TransactionID, pion.BindingRequest)
	request.Add(pion.AttrChangeRequest, []byte{0x00, 0x00, 0x00, 0x02})

	_, err = mapTestConn.roundTrip(request, mapTestConn.RemoteAddr)
	if err == nil {
		return NatBehavior_NAT_BEHAVIOR_ADDRESS_DEPENDENT, nil
	} else if errors.Is(err, errTimedOut) {
		return NatBehavior_NAT_BEHAVIOR_ADDRESS_AND_PORT_DEPENDENT, nil
	}

	return NatBehavior_NAT_BEHAVIOR_UNKNOWN, fmt.Errorf("unexpected error: %w", err)
}

type stunServerConn struct {
	conn        net.PacketConn
	LocalAddr   net.Addr
	RemoteAddr  *net.UDPAddr
	OtherAddr   *net.UDPAddr
	messageChan chan *pion.Message
}

func (c *stunServerConn) Close() error {
	return c.conn.Close()
}

// Send request and wait for response or timeout
func (c *stunServerConn) roundTrip(msg *pion.Message, addr net.Addr) (*pion.Message, error) {
	_ = msg.NewTransactionID()
	_, err := c.conn.WriteTo(msg.Raw, addr)
	if err != nil {
		return nil, err
	}

	// Wait for response or timeout
	select {
	case m, ok := <-c.messageChan:
		if !ok {
			return nil, errResponseMessage
		}
		return m, nil
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil, errTimedOut
	}
}

// Given an address string, returns a StunServerConn
func connect(ctx context.Context, addrStr string) (*stunServerConn, error) {
	addr, err := net.ResolveUDPAddr("udp4", addrStr)
	if err != nil {
		return nil, err
	}

	c, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return nil, err
	}

	mChan := listen(ctx, c)

	return &stunServerConn{
		conn:        c,
		LocalAddr:   c.LocalAddr(),
		RemoteAddr:  addr,
		messageChan: mChan,
	}, nil
}

// taken from https://github.com/pion/stun/blob/master/cmd/stun-traversal/main.go
func listen(ctx context.Context, conn *net.UDPConn) (messages chan *pion.Message) {
	messages = make(chan *pion.Message)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			buf := make([]byte, 1024)

			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				close(messages)
				return
			}
			buf = buf[:n]

			m := new(pion.Message)
			m.Raw = buf
			err = m.Decode()
			if err != nil {
				close(messages)
				return
			}

			messages <- m
		}
	}()
	return
}

// Parse a STUN message
func parse(msg *pion.Message) (ret struct {
	xorAddr    *pion.XORMappedAddress
	otherAddr  *pion.OtherAddress
	respOrigin *pion.ResponseOrigin
	mappedAddr *pion.MappedAddress
	software   *pion.Software
},
) {
	ret.mappedAddr = &pion.MappedAddress{}
	ret.xorAddr = &pion.XORMappedAddress{}
	ret.respOrigin = &pion.ResponseOrigin{}
	ret.otherAddr = &pion.OtherAddress{}
	ret.software = &pion.Software{}
	if ret.xorAddr.GetFrom(msg) != nil {
		ret.xorAddr = nil
	}
	if ret.otherAddr.GetFrom(msg) != nil {
		ret.otherAddr = nil
	}
	if ret.respOrigin.GetFrom(msg) != nil {
		ret.respOrigin = nil
	}
	if ret.mappedAddr.GetFrom(msg) != nil {
		ret.mappedAddr = nil
	}
	if ret.software.GetFrom(msg) != nil {
		ret.software = nil
	}
	for _, attr := range msg.Attributes {
		switch attr.Type {
		case
			pion.AttrXORMappedAddress,
			pion.AttrOtherAddress,
			pion.AttrResponseOrigin,
			pion.AttrMappedAddress,
			pion.AttrSoftware:
		default:
		}
	}
	return ret
}
