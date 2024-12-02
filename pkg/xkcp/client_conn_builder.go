package xkcp

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/xtaci/kcptun/std"
	"net"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"
)

// wrap smux
func BuildClientSmuxSession(conf *SmuxConf, kcpconn *kcp.UDPSession) (*smux.Session, error) {
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

func BuildClientKCP(config *Config, local, remote string) (*kcp.UDPSession, error) {
	localAddr, err := net.ResolveUDPAddr("udp", local)
	if err != nil {
		return nil, err
	}

	remoteAddr, err := net.ResolveUDPAddr("udp", remote)
	if err != nil {
		return nil, err
	}

	localUdpConn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, err
	}

	var convid uint32
	if err := binary.Read(rand.Reader, binary.LittleEndian, &convid); err != nil {
		_ = localUdpConn.Close()
		return nil, err
	}

	block := GetBlockCrypt(config.Key)
	kcpconn, err := kcp.NewConn4(convid, remoteAddr, block, config.FECConf.DataShard, config.FECConf.ParityShard, true, localUdpConn)
	if err != nil {
		return nil, errors.Wrap(err, "BuildClientKCP.NewConn4()")
	}

	kcpconn.SetWriteDelay(false)
	if config.ModeConf != nil {
		kcpconn.SetNoDelay(config.ModeConf.NoDelay, config.ModeConf.Interval, config.ModeConf.Resend, config.ModeConf.NoCongestion)
	}

	kcpconn.SetWindowSize(config.SndWnd, config.RcvWnd)
	kcpconn.SetMtu(config.MTU)
	kcpconn.SetACKNoDelay(config.AckNodelay)

	if err := kcpconn.SetDSCP(config.DSCP); err != nil {
		log.Errorf("set dscp error: %v", err)
	}

	if err := kcpconn.SetReadBuffer(config.SockBuf); err != nil {
		log.Errorf("set read buffer error: %v", err)
	}

	if err := kcpconn.SetWriteBuffer(config.SockBuf); err != nil {
		log.Errorf("set write buffer error: %v", err)
	}

	// handshake
	version := config.Version()
	if _, err := kcpconn.Write([]byte(version)); err != nil {
		_ = kcpconn.Close()
		return nil, err
	}

	_ = kcpconn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// wait for handshake
	buf := make([]byte, 1024)
	n, err := kcpconn.Read(buf)
	if err != nil {
		_ = kcpconn.Close()
		return nil, err
	}

	_ = kcpconn.SetReadDeadline(time.Time{})

	if string(buf[:n]) != version {
		_ = kcpconn.Close()
		log.Errorf("handshake failed, remote: %s, local: %s", string(buf), version)
		return nil, errors.New("handshake failed")
	}

	return kcpconn, nil
}
