// vim: tabstop=4 softtabstop=4 shiftwidth=4 noexpandtab tw=72

package main

import (
	"github.com/ragzilla/telepoller"
	"github.com/ragzilla/telepoller/telepoller_snmp/snmp"
)

func main() {
	s := snmp.NewSnmp()
	f := telepoller.NewFramework()
	if err := f.Init(s); err != nil {
		panic(err)
	}
	if err := f.Run(); err != nil {
		panic(err)
	}
	f.Done()
	return
}
