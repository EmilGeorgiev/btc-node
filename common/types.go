package common

import "fmt"

type Addr struct {
	IP   string
	Port int64
}

func (a Addr) String() string {
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}
