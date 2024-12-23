package xkcp

import (
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/xtaci/smux"
	"sync"
)

type NewConnObserver interface {
	OnNewConn(target string, conn *smux.Session, isServer bool)
}

type ConnectRequest struct {
	LocalAddr  string `json:"local_addr"`
	RemoteAddr string `json:"remote_addr"`
	LocalPort  int    `json:"local_port"`
	RemotePort int    `json:"remote_port"`
	IsActive   bool   `json:"is_active"`
}

type ConnBuilder struct {
	uniqueFilter sync.Map
	observers    []NewConnObserver
}

func NewConnBuilder() *ConnBuilder {
	return &ConnBuilder{}
}

// notify new connection connect success
func (b *ConnBuilder) notifyNewConn(target string, conn *smux.Session, isServer bool) {
	for _, observer := range b.observers {
		observer.OnNewConn(target, conn, isServer)
	}
}

func (b *ConnBuilder) AddObserver(observer NewConnObserver) {
	b.observers = append(b.observers, observer)
}

func (b *ConnBuilder) Build(kcpConf *Config, target string, req *ConnectRequest) error {
	if _, ok := b.uniqueFilter.LoadOrStore(target, 1); ok {
		return fmt.Errorf("connecting")
	}

	defer b.uniqueFilter.Delete(target)

	var smuxSession *smux.Session
	var isServer bool

	if req.IsActive {
		kcpConn, err := BuildClientKCP(kcpConf, fmt.Sprintf(":%d", req.LocalPort), fmt.Sprintf("%s:%d", req.RemoteAddr, req.RemotePort))
		if err != nil {
			log.Errorf("build kcp conn error: %v", err)
			return err
		}

		smuxSession, err = BuildClientSmuxSession(kcpConf.SmuxConf, kcpConn)
		if err != nil {
			log.Errorf("build smux session error: %v", err)
			return err
		}

		isServer = false

	} else {
		kcpConn, err := BuildServerKCP(kcpConf, fmt.Sprintf(":%d", req.LocalPort))
		if err != nil {
			log.Errorf("build kcp conn error: %v", err)
			return err
		}

		smuxSession, err = BuildServerSmuxSession(kcpConf.SmuxConf, kcpConn)
		if err != nil {
			log.Errorf("build smux session error: %v", err)
			return err
		}

		isServer = true
	}

	b.notifyNewConn(target, smuxSession, isServer)
	return nil
}
