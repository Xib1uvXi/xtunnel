package xkcp

import (
	"github.com/stretchr/testify/require"
	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"
	"sync"
	"testing"
	"time"
)

func Test_buildServerKCP(t *testing.T) {
	remote := "127.0.0.1:17888"

	var wg sync.WaitGroup

	var serverConn *kcp.UDPSession

	defer func() {
		if serverConn != nil {
			_ = serverConn.Close()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		conf := DefaultConf
		sc, err := BuildServerKCP(conf, remote)
		require.NoError(t, err)
		require.NotNil(t, sc)

		serverConn = sc
	}()

	time.Sleep(time.Second)

	local := "127.0.0.1:16888"

	conf := DefaultConf
	clientConn, err := BuildClientKCP(conf, local, remote)
	require.NoError(t, err)
	require.NotNil(t, clientConn)

	defer clientConn.Close()

	require.NoError(t, err)

	wg.Wait()
}

func Test_buildServerSmuxSession(t *testing.T) {
	remote := "127.0.0.1:19888"

	var wg sync.WaitGroup

	var serverConn *kcp.UDPSession
	var serverSmux *smux.Session

	defer func() {
		if serverSmux != nil {
			_ = serverSmux.Close()
		}

		if serverConn != nil {
			_ = serverConn.Close()
		}

	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		conf := DefaultConf
		sc, err := BuildServerKCP(conf, remote)
		require.NoError(t, err)
		require.NotNil(t, sc)
		serverConn = sc

		ss, err := BuildServerSmuxSession(DefaultConf.SmuxConf, sc)
		require.NoError(t, err)
		require.NotNil(t, ss)
		serverSmux = ss
	}()

	time.Sleep(time.Second)

	local := "127.0.0.1:15888"

	conf := DefaultConf
	clientConn, err := BuildClientKCP(conf, local, remote)
	require.NoError(t, err)
	require.NotNil(t, clientConn)

	defer clientConn.Close()

	clientSmux, err := BuildClientSmuxSession(DefaultConf.SmuxConf, clientConn)
	require.NoError(t, err)
	require.NotNil(t, clientSmux)
	defer clientSmux.Close()

	wg.Wait()
}
