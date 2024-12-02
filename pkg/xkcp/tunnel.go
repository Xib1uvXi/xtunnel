package xkcp

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/kcptun/std"
	"github.com/xtaci/smux"
	"io"
	"net"
	"sync"
	"time"
)

const (
	scavengettl    = 60 * 60
	scavengePeriod = 5
)

type timedSession struct {
	ctx        context.Context
	cancel     context.CancelFunc
	tcpListen  *net.TCPListener
	target     string
	session    *smux.Session
	expiryDate time.Time
}

type Tunnel struct {
	ctx         context.Context
	cancel      context.CancelFunc
	locker      sync.Mutex
	portMap     map[string]int
	chScavenger chan timedSession
	proxyTarget string
}

func NewTunnel(proxyTarget string) *Tunnel {
	ctx, cancel := context.WithCancel(context.Background())
	tunnel := &Tunnel{
		ctx:         ctx,
		cancel:      cancel,
		portMap:     make(map[string]int),
		chScavenger: make(chan timedSession, 128),
		proxyTarget: proxyTarget,
	}

	return tunnel
}

func (k *Tunnel) Close() {
	k.cancel()
}

func (k *Tunnel) KcpState() *kcp.Snmp {
	return kcp.DefaultSnmp.Copy()
}

func (k *Tunnel) FindPort(target string) (int, bool) {
	k.locker.Lock()
	defer k.locker.Unlock()

	port, ok := k.portMap[target]
	return port, ok
}

func (k *Tunnel) OnNewConn(target string, conn *smux.Session) {
	k.locker.Lock()
	defer k.locker.Unlock()

	tcpListen, err := net.ListenTCP("tcp", nil)
	if err != nil {
		log.Errorf("listen tcp error: %v", err)
		return
	}

	tcpListenPort := tcpListen.Addr().(*net.TCPAddr).Port

	ctx, cancel := context.WithCancel(k.ctx)
	go k.handleOut(target, conn)
	go k.handleIn(target, tcpListen, conn)

	k.portMap[target] = tcpListenPort

	k.chScavenger <- timedSession{
		ctx:        ctx,
		cancel:     cancel,
		tcpListen:  tcpListen,
		target:     target,
		session:    conn,
		expiryDate: time.Now().Add(time.Duration(scavengettl) * time.Second),
	}

	log.Infof("new conn: %s, port: %d", target, tcpListenPort)
}

func (k *Tunnel) scavenger() {
	ticker := time.NewTicker(scavengePeriod * time.Second)
	defer ticker.Stop()
	var sessionList []timedSession
	for {
		select {
		case <-k.ctx.Done():
			k.locker.Lock()
			for i := range sessionList {
				s := sessionList[i]
				s.cancel()
				_ = s.session.Close()
				s.tcpListen.Close()
			}
			k.locker.Unlock()

			log.Infof("kcp tunnel scavenger exit")
			return

		case item := <-k.chScavenger:
			sessionList = append(sessionList, timedSession{
				cancel:     item.cancel,
				tcpListen:  item.tcpListen,
				target:     item.target,
				session:    item.session,
				expiryDate: item.expiryDate.Add(time.Duration(scavengettl) * time.Second),
			})

		case <-ticker.C:
			var newList []timedSession
			for i := range sessionList {
				s := sessionList[i]
				if s.session.IsClosed() {
					log.Infof("scavenger: session normally closed: %s", s.target)
					s.cancel()
					s.tcpListen.Close()
					k.locker.Lock()
					delete(k.portMap, s.target)
					k.locker.Unlock()

				} else if time.Now().After(s.expiryDate) {
					_ = s.session.Close()
					s.tcpListen.Close()
					s.cancel()
					k.locker.Lock()
					delete(k.portMap, s.target)
					k.locker.Unlock()
					log.Infof("scavenger: session closed due to ttl: %s", s.target)
				} else {
					newList = append(newList, sessionList[i])
				}
			}
			sessionList = newList
		}
	}
}

func (k *Tunnel) handleOut(target string, mux *smux.Session) {
	defer mux.Close()

	for {
		stream, err := mux.AcceptStream()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Errorf("accept stream error: %v", err)
			}
			return
		}

		go func(p1 *smux.Stream) {
			var p2 net.Conn
			var err error

			p2, err = net.Dial("tcp", k.proxyTarget)
			if err != nil {
				log.Errorf("dial to ProxyTarget error: %v", err)
				p1.Close()
				return
			}

			defer p1.Close()
			defer p2.Close()

			select {
			case <-p1.GetDieCh():
				return
			default:
			}

			var s1, s2 io.ReadWriteCloser = p1, p2
			err1, err2 := std.Pipe(s1, s2, 1)
			if err1 != nil && err1 != io.EOF {
				log.Debugf("pipe error: %v, in: %s", err1, target)
			}

			if err2 != nil && err2 != io.EOF {
				log.Debugf("pipe error: %v, in: %s", err2, target)
			}

		}(stream)
	}

}

// handleIn
func (k *Tunnel) handleIn(target string, listener *net.TCPListener, session *smux.Session) {
	for {
		p1, err := listener.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Errorf("tcp <-> %s accept error: %v", target, err)
			}
			return
		}

		go func(session *smux.Session, p1 net.Conn) {
			// handles transport layer
			defer p1.Close()
			p2, err := session.OpenStream()
			if err != nil {
				log.Errorf("open stream error: %v", err)
				return
			}
			defer p2.Close()

			log.Debugf("tcp <-> %s", target)
			var s1, s2 io.ReadWriteCloser = p1, p2
			// stream layer
			err1, err2 := std.Pipe(s1, s2, 0)

			if err1 != nil && err1 != io.EOF {
				log.Debugf("pipe error: %v,  out: %s", err1, target)
			}

			if err2 != nil && err2 != io.EOF {
				log.Debugf("pipe error: %v, out: %s", err2, target)
			}

		}(session, p1)
	}
}
