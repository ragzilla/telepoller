// vim: tabstop=4 softtabstop=4 shiftwidth=4 noexpandtab tw=72

package main

import (
	// "github.com/davecgh/go-spew/spew"
	"github.com/ragzilla/telepoller"
	"github.com/ragzilla/telepoller/telepoller_snmp/snmp"
)

/*
import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	// "github.com/davecgh/go-spew/spew"
	client "github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/toml"
	tsnmp "github.com/ragzilla/telepoller/telepoller_snmp/snmp"
)
*/

func main() {
	s := snmp.NewSnmp()
	f := telepoller.NewFramework()
	if err := f.Init(s); err != nil {
		panic(err)
	}
	// spew.Dump(f)
	if err := f.Run(); err != nil {
		panic(err)
	}
	f.Done()
	return
	/*
		wg := &sync.WaitGroup{}

		for _, a := range snmp.Agents {
			wg.Add(1)
			t := snmp.GetTable("ifMIB")
			if t == nil {
				panic(fmt.Sprintf("table %s not found", "ifMIB"))
			}
			go func(t *tsnmp.Table, a string, c string) {
				defer wg.Done()
				rt, err := t.Build(a, c)
				if err != nil {
					panic("foo")
				}
				for _, rtr := range rt.Rows {
					if len(rtr.Fields) == 0 {
						continue
					}
					rtr.Tags["agent_host"] = a
					pt, err := client.NewPoint("ifMIB", rtr.Tags, rtr.Fields, rt.Time)
					if err != nil {
						panic(err)
					}
					fmt.Println(pt.String())
				}
			}(t, a, snmp.Community)
		}
		wg.Wait()
	*/
}
