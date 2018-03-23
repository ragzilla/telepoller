// vim: ts=4:sw=4

package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/influxdata/toml"
	tsnmp "github.com/ragzilla/telepoller/snmp"
	"io/ioutil"
	"os"
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
	spew.Dump(snmp)
}
