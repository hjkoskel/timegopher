package timesync

import (
	"fmt"
	"strings"
	"time"

	"math/rand"

	"github.com/beevik/ntp"
)

type NtpSync struct {
	Servers      []string
	QueryTimeout time.Duration
}

func GetDefaultFinnishNTP() NtpSync {
	return NtpSync{
		Servers:      []string{"0.fi.pool.ntp.org", "1.fi.pool.ntp.org", "2.fi.pool.ntp.org", "3.fi.pool.ntp.org"},
		QueryTimeout: time.Second * 30,
	}
}

func (p *NtpSync) pickServerList() []string {
	shuffled := make([]string, len(p.Servers))
	copy(shuffled, p.Servers)
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return shuffled
}

func (p *NtpSync) GetOffset() (time.Duration, error) {
	if p.QueryTimeout < time.Millisecond*100 {
		p.QueryTimeout = time.Second * 30
	}

	errList := []string{}
	lst := p.pickServerList()
	for i, name := range lst {
		resp, err := ntp.QueryWithOptions(name, ntp.QueryOptions{Timeout: p.QueryTimeout})
		if err != nil {
			errList = append(errList, fmt.Sprintf("server:%v name:%s error: %s", i, name, err))
			continue
		}
		errvalid := resp.Validate()
		if errvalid == nil {
			return resp.ClockOffset, nil
		}
		errList = append(errList, fmt.Sprintf("server:%v name:%s invalid: %s", i, name, errvalid))
	}

	return 0, fmt.Errorf("failed NTP servers [%s]", strings.Join(errList, ","))
}
