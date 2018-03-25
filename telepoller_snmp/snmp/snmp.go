// vim: tabstop=4 softtabstop=4 shiftwidth=4 noexpandtab tw=72
// The contents of this file are Copyright (c) 2015 InfluxDB
// Reproduced under the terms of The MIT License (MIT)

package snmp

import (
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	radix "github.com/armon/go-radix"
	"github.com/influxdata/toml"
	"github.com/ragzilla/telepoller"
	"github.com/soniah/gosnmp"
)

// Snmp holds the configuration for the agent
type Snmp struct {
	framework      *telepoller.TpFramework
	Timeout        time.Duration
	Retries        int
	MaxRepetitions uint8
	Tables         []Table `toml:"table"`
}

func NewSnmp() *Snmp {
	s := Snmp{
		Retries:        3,
		Timeout:        5 * time.Second,
		MaxRepetitions: 10,
	}
	return &s
}

func (s *Snmp) Init(framework *telepoller.TpFramework, config string) error {
	s.framework = framework

	// load configuration
	f, err := os.Open(config)
	if err != nil {
		return err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	if err := toml.Unmarshal(buf, s); err != nil {
		return err
	}

	// initialize filters
	for idx, _ := range s.Tables {
		if err := s.Tables[idx].Init(s); err != nil {
			return err
		}
	}
	// fmt.Println("initialized snmp!")
	// spew.Dump(s)
	return nil
}

func (s *Snmp) NewJob(j *telepoller.TpJob, cb func()) {
	// fmt.Println("Snmp.NewJob() new job:", j)
	// check we have table and community params
	if _, ok := j.Params["community"]; !ok {
		cb()
		return
	}
	if _, ok := j.Params["table"]; !ok {
		cb()
		return
	}
	table := s.GetTable(j.Params["table"])
	if table == nil {
		fmt.Printf("Bad table \"%v\" in request %v\n", j.Params["table"], j)
		cb()
		return
	}
	community := j.Params["community"]
	if community == "" {
		fmt.Printf("No community in request %v\n", j)
		cb()
		return
	}
	for k, v := range j.Hosts {
		if k == "" || v == "" {
			continue
		}
		s.framework.Publish(nil)
	}
	cb()
	return
}

func (s *Snmp) GetTable(table string) *Table {
	for idx, _ := range s.Tables {
		if table == s.Tables[idx].Name {
			return &s.Tables[idx]
		}
	}
	return nil
}

// Table holds the configuration for a SNMP table.
type Table struct {
	// Name will be the name of the measurement.
	Name string
	// Fields is the tags and values to look up.
	Fields  []Field  `toml:"field"`
	Filters []Filter `toml:"filter"`
}

func (t *Table) Init(s *Snmp) error {
	for idx, _ := range t.Filters {
		if err := t.Filters[idx].Init(t); err != nil {
			return err
		}
	}
	return nil
}

func (t *Table) GetField(field string) *Field {
	for idx, _ := range t.Fields {
		if field == t.Fields[idx].Name {
			return &t.Fields[idx]
		}
	}
	return nil
}

// Build retrieves fields specified in a table and returns an RTable
func (t Table) Build(agent string, community string) (*RTable, error) {
	rows := map[string]RTableRow{}
	rl := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for _, f := range t.Fields {
		wg.Add(1)
		go func(f Field, agent string, community string) error {
			defer wg.Done()
			if len(f.Oid) == 0 {
				panic(fmt.Sprintf("cannot have empty OID on field %s", f.Name))
			}
			var oid string
			if f.Oid[0] == '.' {
				oid = f.Oid
			} else {
				// make sure OID has "." because the BulkWalkAll results do, and the prefix needs to match
				oid = "." + f.Oid
			}

			// ifv contains a mapping of table OID index to field value
			ifv := map[string]interface{}{}

			gs := gosnmpWrapper{&gosnmp.GoSNMP{}}
			gs.Target = agent
			gs.Port = 161
			gs.Version = gosnmp.Version2c
			gs.Community = community
			gs.MaxRepetitions = 10
			gs.Retries = 3
			gs.Timeout = 5 * time.Second
			if err := gs.Connect(); err != nil {
				return Errorf(err, "setting up connection")
			}

			err := gs.Walk(oid, func(ent gosnmp.SnmpPDU) error {
				if len(ent.Name) <= len(oid) || ent.Name[:len(oid)+1] != oid+"." {
					return NestedError{} // break the walk
				}

				idx := ent.Name[len(oid):]
				if f.OidIndexSuffix != "" {
					if !strings.HasSuffix(idx, f.OidIndexSuffix) {
						// this entry doesn't match our OidIndexSuffix. skip it
						return nil
					}
					idx = idx[:len(idx)-len(f.OidIndexSuffix)]
				}

				fv, err := fieldConvert(f.Conversion, ent.Value)
				if err != nil {
					return Errorf(err, "converting %q (OID %s) for field %s", ent.Value, ent.Name, f.Name)
				}
				ifv[idx] = fv
				return nil
			})

			if err != nil {
				if _, ok := err.(NestedError); !ok {
					/*
						return nil, Errorf(err, "performing bulk walk for field %s", f.Name)
					*/
					return Errorf(err, "performing bulk walk for field %s", f.Name)
				}
			}

			for idx, v := range ifv {
				rl.Lock()
				rtr, ok := rows[idx]
				if !ok {
					rtr = RTableRow{}
					rtr.Tags = map[string]string{}
					rtr.Fields = map[string]interface{}{}
					rows[idx] = rtr
				}
				rl.Unlock()
				// don't add an empty string
				if vs, ok := v.(string); !ok || vs != "" {
					if f.IsTag {
						if ok {
							rtr.Tags[f.Name] = vs
						} else {
							rtr.Tags[f.Name] = fmt.Sprintf("%v", v)
						}
					} else {
						rtr.Fields[f.Name] = v
					}
				}
			}
			return nil
		}(f, agent, community)
	}
	wg.Wait()

	rt := RTable{
		Name: t.Name,
		Time: time.Now(), //TODO record time at start
		Rows: make([]RTableRow, 0, len(rows)),
	}
	for _, r := range rows {
		func(r *RTableRow) {
			for idx, _ := range t.Filters {
				if !t.Filters[idx].Check(r.Tags[t.Filters[idx].Name]) {
					return
				}
			}
			rt.Rows = append(rt.Rows, *r)
		}(&r)
	}
	return &rt, nil
}

type Filter struct {
	// the name of the tag to filter on
	Name string `toml:"name"`

	// tag values to include/exclude
	Values []string `toml:"values"`

	// is this a prefix match
	Prefix bool `toml:"prefix"`

	// are we excluding here?
	Exclude bool `toml:"exclude"`

	// map, and default value
	Filter *radix.Tree
}

func (f *Filter) Init(t *Table) error {
	if f.Filter == nil {
		f.Filter = radix.New()
	}
	for _, v := range f.Values {
		f.Filter.Insert(v, true)
	}
	return nil
}

func (f *Filter) Check(c string) bool {
	rv := false
	if f.Prefix {
		_, _, rv = f.Filter.LongestPrefix(c)
	} else {
		_, rv = f.Filter.Get(c)
	}
	return rv != f.Exclude
}

// Field holds the configuration for a Field to look up.
type Field struct {
	// Name will be the name of the field.
	Name string
	// OID is prefix for this field. The plugin will perform a walk through all
	// OIDs with this as their parent. For each value found, the plugin will strip
	// off the OID prefix, and use the remainder as the index. For multiple fields
	// to show up in the same row, they must share the same index.
	Oid string
	// OidIndexSuffix is the trailing sub-identifier on a table record OID that will be stripped off to get the record's index.
	OidIndexSuffix string
	// IsTag controls whether this OID is output as a tag or a value.
	IsTag bool
	// Conversion controls any type conversion that is done on the value.
	//  "float"/"float(0)" will convert the value into a float.
	//  "float(X)" will convert the value into a float, and then move the decimal before Xth right-most digit.
	//  "int" will conver the value into an integer.
	//  "hwaddr" will convert a 6-byte string to a MAC address.
	//  "ipaddr" will convert the value to an IPv4 or IPv6 address.
	Conversion string
}

// RTable is the resulting table built from a Table.
type RTable struct {
	// Name is the name of the field, copied from Table.Name.
	Name string
	// Time is the time the table was built.
	Time time.Time
	// Rows are the rows that were found, one row for each table OID index found.
	Rows []RTableRow
}

// RTableRow is the resulting row containing all the OID values which shared
// the same index.
type RTableRow struct {
	// Tags are all the Field values which had IsTag=true.
	Tags map[string]string
	// Fields are all the Field values which had IsTag=false.
	Fields map[string]interface{}
}

// NestedError wraps an error returned from deeper in the code.
type NestedError struct {
	// Err is the error from where the NestedError was constructed.
	Err error
	// NestedError is the error that was passed back from the called function.
	NestedErr error
}

// Error returns a concatenated string of all the nested errors.
func (ne NestedError) Error() string {
	return ne.Err.Error() + ": " + ne.NestedErr.Error()
}

// Errorf is a convenience function for constructing a NestedError.
func Errorf(err error, msg string, format ...interface{}) error {
	return NestedError{
		NestedErr: err,
		Err:       fmt.Errorf(msg, format...),
	}
}

// gosnmpWrapper wraps a *gosnmp.GoSNMP object so we can use it as a snmpConnection.
type gosnmpWrapper struct {
	*gosnmp.GoSNMP
}

// Host returns the value of GoSNMP.Target.
func (gsw gosnmpWrapper) Host() string {
	return gsw.Target
}

// Walk GoSNMP.BulkWalk()
// Also, if any error is encountered, it will just once reconnect and try again.
func (gsw gosnmpWrapper) Walk(oid string, fn gosnmp.WalkFunc) error {
	var err error
	// On error, retry once.
	// Unfortunately we can't distinguish between an error returned by gosnmp, and one returned by the walk function.
	for i := 0; i < 2; i++ {
		err = gsw.GoSNMP.BulkWalk(oid, fn)
		if err == nil {
			return nil
		}
		if err := gsw.GoSNMP.Connect(); err != nil {
			return Errorf(err, "reconnecting")
		}
	}
	return err
}

// fieldConvert converts from any type according to the conv specification
//  "float"/"float(0)" will convert the value into a float.
//  "float(X)" will convert the value into a float, and then move the decimal before Xth right-most digit.
//  "int" will convert the value into an integer.
//  "hwaddr" will convert the value into a MAC address.
//  "ipaddr" will convert the value into into an IP address.
//  "" will convert a byte slice into a string.
func fieldConvert(conv string, v interface{}) (interface{}, error) {
	if conv == "" {
		if bs, ok := v.([]byte); ok {
			return string(bs), nil
		}
		return v, nil
	}

	var d int
	if _, err := fmt.Sscanf(conv, "float(%d)", &d); err == nil || conv == "float" {
		switch vt := v.(type) {
		case float32:
			v = float64(vt) / math.Pow10(d)
		case float64:
			v = float64(vt) / math.Pow10(d)
		case int:
			v = float64(vt) / math.Pow10(d)
		case int8:
			v = float64(vt) / math.Pow10(d)
		case int16:
			v = float64(vt) / math.Pow10(d)
		case int32:
			v = float64(vt) / math.Pow10(d)
		case int64:
			v = float64(vt) / math.Pow10(d)
		case uint:
			v = float64(vt) / math.Pow10(d)
		case uint8:
			v = float64(vt) / math.Pow10(d)
		case uint16:
			v = float64(vt) / math.Pow10(d)
		case uint32:
			v = float64(vt) / math.Pow10(d)
		case uint64:
			v = float64(vt) / math.Pow10(d)
		case []byte:
			vf, _ := strconv.ParseFloat(string(vt), 64)
			v = vf / math.Pow10(d)
		case string:
			vf, _ := strconv.ParseFloat(vt, 64)
			v = vf / math.Pow10(d)
		}
		return v, nil
	}

	if conv == "int" {
		switch vt := v.(type) {
		case float32:
			v = int64(vt)
		case float64:
			v = int64(vt)
		case int:
			v = int64(vt)
		case int8:
			v = int64(vt)
		case int16:
			v = int64(vt)
		case int32:
			v = int64(vt)
		case int64:
			v = int64(vt)
		case uint:
			v = uint64(vt)
		case uint8:
			v = uint64(vt)
		case uint16:
			v = uint64(vt)
		case uint32:
			v = uint64(vt)
		case uint64:
			v = uint64(vt)
		case []byte:
			v, _ = strconv.Atoi(string(vt))
		case string:
			v, _ = strconv.Atoi(vt)
		}
		return v, nil
	}

	if conv == "hwaddr" {
		switch vt := v.(type) {
		case string:
			v = net.HardwareAddr(vt).String()
		case []byte:
			v = net.HardwareAddr(vt).String()
		default:
			return nil, fmt.Errorf("invalid type (%T) for hwaddr conversion", v)
		}
		return v, nil
	}

	if conv == "ipaddr" {
		var ipbs []byte

		switch vt := v.(type) {
		case string:
			ipbs = []byte(vt)
		case []byte:
			ipbs = vt
		default:
			return nil, fmt.Errorf("invalid type (%T) for ipaddr conversion", v)
		}

		switch len(ipbs) {
		case 4, 16:
			v = net.IP(ipbs).String()
		default:
			return nil, fmt.Errorf("invalid length (%d) for ipaddr conversion", len(ipbs))
		}

		return v, nil
	}

	return nil, fmt.Errorf("invalid conversion type '%s'", conv)
}
