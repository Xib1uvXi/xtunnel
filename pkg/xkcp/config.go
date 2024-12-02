package xkcp

import (
	"fmt"
	"github.com/goccy/go-json"
	"os"
)

type Config struct {
	Key        string    `json:"key"`
	MTU        int       `json:"mtu"`
	SndWnd     int       `json:"sndwnd"`
	RcvWnd     int       `json:"rcvwnd"`
	DSCP       int       `json:"dscp"`
	AckNodelay bool      `json:"acknodelay"`
	SockBuf    int       `json:"sockbuf"`
	ModeConf   *ModeConf `json:"mode"`
	FECConf    *FECConf  `json:"fec"`
	SmuxConf   *SmuxConf `json:"smux"`
}

func ParseJSONConfig(config *Config, path string) error {
	file, err := os.Open(path) // For read access.
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(config)
}

// Identical parameters
func (c *Config) Version() string {
	str := fmt.Sprintf("%s_%v_%d", c.Key, c.SmuxConf.NoComp, c.SmuxConf.SmuxVer)
	return str
}

type FECConf struct {
	DataShard   int `json:"datashard"`
	ParityShard int `json:"parityshard"`
}

type ModeConf struct {
	NoDelay      int `json:"nodelay"`
	Interval     int `json:"interval"`
	Resend       int `json:"resend"`
	NoCongestion int `json:"nc"`
}

type SmuxConf struct {
	SmuxBuf   int  `json:"smuxbuf"`
	StreamBuf int  `json:"streambuf"`
	SmuxVer   int  `json:"smuxver"`
	KeepAlive int  `json:"keepalive"`
	NoComp    bool `json:"nocomp"`
}

var DefaultConf = &Config{
	Key:        "test",
	MTU:        1200,
	SndWnd:     1024,
	RcvWnd:     1024 * 4,
	DSCP:       46,
	AckNodelay: false,
	SockBuf:    16777217,
	ModeConf:   &ModeConf{NoDelay: 0, Interval: 30, Resend: 2, NoCongestion: 1},
	FECConf:    &FECConf{DataShard: 10, ParityShard: 3},
	SmuxConf: &SmuxConf{
		SmuxBuf:   16777217,
		StreamBuf: 8388608,
		SmuxVer:   2,
		KeepAlive: 10,
		NoComp:    false,
	},
}
