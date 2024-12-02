package netutil

import "net"

func IsSameSegmentIP(ip1, ip2 string, bc int) bool {
	ip1Bytes := net.ParseIP(ip1).To4()
	ip2Bytes := net.ParseIP(ip2).To4()

	if ip1Bytes == nil || ip2Bytes == nil {
		return false // 非法IP地址
	}

	for i := 0; i < bc; i++ {
		if ip1Bytes[i] != ip2Bytes[i] {
			return false
		}
	}
	return true
}
