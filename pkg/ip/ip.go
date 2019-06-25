package ip

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func Check(ip string) error {
	host := strings.Split(ip, ":")

	if ip := net.ParseIP(host[0]); ip == nil {
		return fmt.Errorf("ip error %s", host[0])
	}

	if len(host) == 2 {
		if _, err := strconv.Atoi(host[1]); err != nil {
			return fmt.Errorf("port error %s", host[1])
		}
	}

	return nil
}
