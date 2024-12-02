package xkcp

import (
	"crypto/sha1"
	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
)

const salt = "kcp-go"

func GetBlockCrypt(seed string) kcp.BlockCrypt {
	pass := pbkdf2.Key([]byte(seed), []byte(salt), 4096, 32, sha1.New)
	//block, err := kcp.NewAESBlockCrypt(pass)
	block, err := kcp.NewSalsa20BlockCrypt(pass)
	if err != nil {
		panic(err)
	}

	return block
}
