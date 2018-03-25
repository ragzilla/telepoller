// vim: tabstop=4 softtabstop=4 shiftwidth=4 noexpandtab tw=72

package main

import (
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
	f.Init(s)
	/*
		f, err := os.Open("telepoller_snmp.conf")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		buf, err := ioutil.ReadAll(f)
		if err != nil {
			panic(err)
		}
		snmp := tsnmp.NewSnmp()
		if err := toml.Unmarshal(buf, &snmp); err != nil {
			panic(err)
		}
		if err := snmp.Init(); err != nil {
			panic(err)
		}

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
