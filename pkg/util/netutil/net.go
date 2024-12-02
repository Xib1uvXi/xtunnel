package netutil

import (
	"fmt"
	"github.com/Doraemonkeys/reliableUDP"
	"github.com/Xib1uvXi/xtunnel/pkg/util/json"
	"github.com/go-kratos/kratos/v2/log"
	"net"
	"time"
)

const (
	maxPayloadSize = 1024 * 1024
)

func ConnSendMessage(conn net.Conn, data interface{}) error {
	payload, err := json.StringifyJsonToBytesWithErr(data)
	if err != nil {
		log.Errorf("stringify json error: %v", err)
		return err
	}

	_, err = conn.Write(payload)
	if err != nil {
		log.Errorf("write message error: %v", err)
		return err
	}

	return nil
}

func ConnReceiveMessage(conn net.Conn, data interface{}) error {
	payload := make([]byte, maxPayloadSize)
	n, err := conn.Read(payload)
	if err != nil {
		log.Errorf("read message error: %v", err)
		return err
	}

	err = json.ParseJsonFromBytes(payload[:n], data)
	if err != nil {
		log.Errorf("%s parse json error: %v", string(payload[:n]), err)
		return err
	}

	return nil
}

func UDPRandListen() (*net.UDPConn, error) {
	randPort := RandPort()
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", randPort))

	if err != nil {
		return nil, err
	}

	udpConn, err := net.ListenUDP("udp4", addr)

	if err != nil {
		return nil, err
	}

	return udpConn, nil
}

func TCPRandListen() (*net.TCPListener, error) {
	randPort := RandPort()
	addr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", randPort))

	if err != nil {
		return nil, err
	}

	tcpConn, err := net.ListenTCP("tcp4", addr)

	if err != nil {
		return nil, err
	}

	return tcpConn, nil
}

// 发送超时时间为timeout,如果timeout为0则默认为4秒
func RUDPSendMessage(conn *reliableUDP.ReliableUDP, addr string, msg interface{}, timeout time.Duration) error {
	data, err := json.StringifyJsonToBytesWithErr(msg)
	if err != nil {
		return err
	}

	raddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return err
	}
	err = conn.Send(raddr, data, timeout)
	if err != nil {
		return err
	}
	return nil
}

func RUDPSendUnreliableMessage(conn *reliableUDP.ReliableUDP, addr string, msg interface{}) error {
	data, err := json.StringifyJsonToBytesWithErr(msg)
	if err != nil {
		return err
	}

	raddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return err
	}
	err = conn.SendUnreliable(data, raddr)
	if err != nil {
		return err
	}
	return nil
}

func RUDPReceiveAllMessage(conn *reliableUDP.ReliableUDP, timeout time.Duration, msg interface{}) (*net.UDPAddr, error) {
	data, addr, err := conn.ReceiveAll(timeout)
	if err != nil {
		return nil, err
	}

	err = json.ParseJsonFromBytes(data, msg)
	if err != nil {
		return nil, err
	}

	return addr, nil
}
