package xkcp

import "testing"

func TestConfig_Version(t *testing.T) {
	conf := &Config{
		Key:      "testseed",
		SmuxConf: &SmuxConf{},
	}

	got := conf.Version()
	t.Logf("Version() = %v", got)
}
