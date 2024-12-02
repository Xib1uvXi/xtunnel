package xkcp

import (
	"github.com/xtaci/kcptun/std"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"
)

// wrap smux
func BuildServerSmuxSession(conf *SmuxConf, kcpconn *kcp.UDPSession) (*smux.Session, error) {
	smuxConfig := smux.DefaultConfig()
	smuxConfig.Version = conf.SmuxVer
	smuxConfig.MaxReceiveBuffer = conf.SmuxBuf
	smuxConfig.MaxStreamBuffer = conf.StreamBuf
	smuxConfig.KeepAliveInterval = time.Duration(conf.KeepAlive) * time.Second

	var session *smux.Session
	var err error

	if conf.NoComp {
		session, err = smux.Client(kcpconn, smuxConfig)
	} else {
		session, err = smux.Client(std.NewCompStream(kcpconn), smuxConfig)
	}

	if err != nil {
		log.Errorf("create smux session error: %v", err)
		return nil, err
	}

	return session, nil
}

func BuildServerKCP(config *Config, local string) (*kcp.UDPSession, error) {
	block := GetBlockCrypt(config.Key)

	listen, err := kcp.ListenWithOptions(local, block, config.FECConf.DataShard, config.FECConf.ParityShard)
	if err != nil {
		log.Errorf("listen kcp failed: %v", err)
		return nil, err
	}

	if err := listen.SetDSCP(config.DSCP); err != nil {
		log.Errorf("set dscp error: %v", err)
	}

	if err := listen.SetReadBuffer(config.SockBuf); err != nil {
		log.Errorf("set read buffer error: %v", err)
	}

	if err := listen.SetWriteBuffer(config.SockBuf); err != nil {
		log.Errorf("set write buffer error: %v", err)
	}

	_ = listen.SetReadDeadline(time.Now().Add(time.Second * 10))

	conn, err := listen.AcceptKCP()
	if err != nil {
		_ = listen.Close()
		log.Errorf("accept kcp failed: %v", err)
		return nil, err
	}

	_ = listen.SetReadDeadline(time.Time{})

	log.Debugf("new client %s<->%s", conn.RemoteAddr(), conn.LocalAddr())

	conn.SetWriteDelay(false)

	if config.ModeConf != nil {
		conn.SetNoDelay(config.ModeConf.NoDelay, config.ModeConf.Interval, config.ModeConf.Resend, config.ModeConf.NoCongestion)
	}

	conn.SetMtu(config.MTU)
	conn.SetWindowSize(config.SndWnd, config.RcvWnd)
	conn.SetACKNoDelay(config.AckNodelay)

	// wait for handshake
	version := config.Version()

	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		_ = conn.Close()
		log.Errorf("read handshake failed: %v", err)
		return nil, err
	}

	_ = conn.SetReadDeadline(time.Time{})

	if string(buf[:n]) != version {
		_ = conn.Close()
		log.Errorf("handshake failed, remote: %s, local: %s", string(buf), version)
		return nil, err
	}

	_, err = conn.Write([]byte(version))
	if err != nil {
		_ = conn.Close()
		log.Errorf("write handshake failed: %v", err)
		return nil, err
	}

	return conn, nil
}
