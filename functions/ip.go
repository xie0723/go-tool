package functions

import (
	"net"
	"strconv"
	"strings"

	"github.com/astaxie/beego/logs"
)

//判断IP是否是本地的IP（IPV4和IPV6）
func IsLocalIP(ip string) (bool, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false, err
	}
	for i := range addrs {
		intf, _, err := net.ParseCIDR(addrs[i].String())
		if err != nil {
			return false, err
		}
		if net.ParseIP(ip).Equal(intf) {
			return true, nil
		}
	}
	return false, nil
}

// IsIntranet 判断是否是内网
func IsIntranet(ip string) bool {
	if !checkIp(ip) {
		return false
	}
	inputIpNum := inetAton(ip)
	innerIpA := inetAton("10.255.255.255")
	innerIpB := inetAton("172.16.255.255")
	innerIpC := inetAton("192.168.255.255")
	innerIpD := inetAton("100.64.255.255")
	innerIpF := inetAton("127.255.255.255")

	return inputIpNum>>24 == innerIpA>>24 || inputIpNum>>20 == innerIpB>>20 ||
		inputIpNum>>16 == innerIpC>>16 || inputIpNum>>22 == innerIpD>>22 ||
		inputIpNum>>24 == innerIpF>>24
}

//检测ip的正确性
func checkIp(ipStr string) bool {
	address := net.ParseIP(ipStr)
	if address == nil {
		logs.Error("ip address is wrong, ip is ", ipStr)
		return false
	}
	return true
}

// ip to int64
func inetAton(ipStr string) int64 {
	bits := strings.Split(ipStr, ".")
	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])
	var sum int64
	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)
	return sum
}
