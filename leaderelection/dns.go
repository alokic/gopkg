package leaderelection

import (
	"github.com/pkg/errors"
	"net"
	"sort"
	"sync"
	"time"
)

/*
  This file implements leader-election based on DNS resolution of a URL.
  How are participants decided?
    All processes having the same URL are the participant.

  Properties for leader election:
  1. Termination: It terminates bound by DNS update latency.
  2. Uniqueness: Only lexicographically highest IP is leader.
  3. Agreement: All participant know and agree that lexicographically highest IP is leader.
*/

type (
	dnsBasedLeader struct {
		done           chan struct{}
		url            string
		myIP           string
		lookupInterval time.Duration
		isLeader       bool
		lastErr        error
		mu             sync.RWMutex
	}
)

// GetOutboundIP gets preferred outbound ip of this machine.
func GetOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80") //udp doesnt need any handshake, so it doesnt ven connect to destination.
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

//Set export CGO_ENABLED=1 and export GODEBUG=netdns=cgo.
//NewDNSBased for DNS based leader election. Returns err if URL not found in DNS.
func NewDNSBased(url, myIP string, lookupInterval time.Duration) Leader {
	d := &dnsBasedLeader{done: make(chan struct{}), url: url, myIP: myIP, lookupInterval: lookupInterval, mu: sync.RWMutex{}}
	go d.run()
	return d
}

func (d *dnsBasedLeader) IsLeader() (bool, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.isLeader, d.lastErr
}

func (d *dnsBasedLeader) Stop() {
	close(d.done)
}

func (d *dnsBasedLeader) run() {
	t := time.NewTicker(d.lookupInterval)
	for {
		select {
		case <-d.done:
			return
		case <-t.C:
			d.setStatus()
		}
	}
}

/*
  there could be small period of time when there is no leader, this is bound by dns update.
  e.g all pods crash and get new IP but DNS is not up-to-date yet.
*/
func (d *dnsBasedLeader) setStatus() {
	ips, err := net.LookupIP(d.url) //TODO make sure it doesn't caches, dnclient_unix.go says it doesn't.
	{
		if err == nil && len(ips) == 0 {
			err = errors.New("no ip found")
		}
		d.mu.Lock()
		d.lastErr = err
		d.mu.Unlock()
	}
	//if there is err in lookup OR no ips leader status is not changed, but err is set, so consult it.
	if err != nil {
		return
	}

	sort.Slice(ips, func(i, j int) bool { return ips[i].String() <= ips[j].String() })
	{
		d.mu.Lock()
		d.isLeader = ips[len(ips)-1].String() == d.myIP
		d.mu.Unlock()
	}
}
