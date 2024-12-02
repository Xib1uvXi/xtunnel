package xkcp

import (
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/require"
	"net"
	"sync"
	"testing"
	"time"
)

func TestTCPListen(t *testing.T) {
	listener, err := net.ListenTCP("tcp", nil)
	require.NoError(t, err)
	defer listener.Close()

	t.Logf(listener.Addr().String())

	port := listener.Addr().(*net.TCPAddr).Port
	t.Logf("port: %d", port)
}

type testTcpServer struct {
	listener *net.TCPListener
}

func newTestTcpServer(addr string) *testTcpServer {
	lisAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", lisAddr)
	if err != nil {
		panic(err)
	}

	s := &testTcpServer{listener: listener}
	go s.accept()
	return s
}

// close
func (s *testTcpServer) close() {
	s.listener.Close()
}

// accept
func (s *testTcpServer) accept() {
	for {
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			return
		}

		go func(conn *net.TCPConn) {
			defer conn.Close()
			for {
				buf := make([]byte, 1024)
				n, err := conn.Read(buf)
				if err != nil {
					return
				}

				if n > 0 {
					log.Infof("recv: %s", string(buf[:n]))
				}

				_, _ = conn.Write(buf[:n])
			}
		}(conn)
	}
}

func TestNewTunnel(t *testing.T) {
	testServer := newTestTcpServer("127.0.0.1:9988")
	defer testServer.close()

	serverTun := NewTunnel("127.0.0.1:9988")
	defer serverTun.Close()

	clientTun := NewTunnel("127.0.0.1:9988")
	defer clientTun.Close()

	remote := "127.0.0.1:29888"
	local := "127.0.0.1:35888"

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		conf := DefaultConf
		sc, err := BuildServerKCP(conf, remote)
		require.NoError(t, err)
		require.NotNil(t, sc)

		ss, err := BuildServerSmuxSession(conf.SmuxConf, sc)
		require.NoError(t, err)
		require.NotNil(t, ss)

		serverTun.OnNewConn("client", ss)
	}()

	time.Sleep(time.Second)

	wg.Add(1)
	go func() {
		defer wg.Done()
		conf := DefaultConf
		clientConn, err := BuildClientKCP(conf, local, remote)
		require.NoError(t, err)
		require.NotNil(t, clientConn)
		clientSmux, err := BuildClientSmuxSession(conf.SmuxConf, clientConn)
		require.NoError(t, err)
		require.NotNil(t, clientSmux)

		clientTun.OnNewConn("server", clientSmux)
	}()

	wg.Wait()

	serverPort, serverReady := clientTun.FindPort("server")
	require.True(t, serverReady)
	require.True(t, serverPort > 0)

	testConn, err := net.Dial("tcp", fmt.Sprintf(":%d", serverPort))
	require.NoError(t, err)
	defer testConn.Close()

	_, err = testConn.Write([]byte("hello"))
	require.NoError(t, err)

	buf := make([]byte, 1024)
	n, err := testConn.Read(buf)
	require.NoError(t, err)
	require.Equal(t, "hello", string(buf[:n]))

	time.Sleep(time.Second)

	//clientPort, clientReady := serverTun.FindPort("client")
	//require.True(t, clientReady)
	//require.True(t, clientPort > 0)
	//
	//testSConn, err := net.Dial("tcp", fmt.Sprintf(":%d", clientPort))
	//require.NoError(t, err)
	//defer testSConn.Close()
	//
	//_, err = testSConn.Write([]byte("hello"))
	//require.NoError(t, err)
	//
	//t.Logf("-------------")
	//
	//bufs := make([]byte, 1024)
	//n, err = testSConn.Read(bufs)
	//require.NoError(t, err)
	//require.Equal(t, "hello", string(bufs[:n]))
	//time.Sleep(time.Second)
}
