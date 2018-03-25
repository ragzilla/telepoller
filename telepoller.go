// vim: ts=4:sw=4

package main

import (
	"fmt"
	"io/ioutil"
	"os"

	_ "github.com/davecgh/go-spew/spew"
	client "github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/toml"
	tsnmp "github.com/ragzilla/telepoller/snmp"
)

func main() {
	f, err := os.Open("telepoller.conf")
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
	// spew.Dump(snmp)
	for _, k := range snmp.Agents {
		fmt.Printf("agent: %s\n", k)
		// snmp.Collect(k, "ifMIB", snmp.Community)
		t := snmp.GetTable("ifMIB")
		rt, err := t.Build(k, snmp.Community)
		if err != nil {
			panic("foo")
		}
		// spew.Dump(rt)
		for _, rtr := range rt.Rows {
			pt, err := client.NewPoint("ifMIB", rtr.Tags, rtr.Fields, rt.Time)
			if err != nil {
				panic(err)
			}
			fmt.Println(pt.String())
		}
	}
}
