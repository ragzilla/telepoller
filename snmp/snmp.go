// vim: ts=4:sw=4
// The contents of this file are Copyright (c) 2015 InfluxDB
// Reproduced under the terms of The MIT License (MIT)

package snmp

import (
	"time"
)

// Snmp holds the configuration for the agent
type Snmp struct {
	// Timeout to wait for a response.
	Timeout time.Duration
	Retries int
	// Values: 1, 2
	Version uint8

	// Parameters for Version 1 & 2
	Community string

	// Parameters for Version 2
	MaxRepetitions uint8

	Tables []Table `toml:"table"`

	// Name & Fields are the elements of a Table.
	// Telegraf chokes if we try to embed a Table. So instead we have to embed the
	// fields of a Table, and construct a Table during runtime.
	Name   string
	Fields []Field `toml:"field"`

	initialized bool
}

// Table holds the configuration for a SNMP table.
type Table struct {
	// Name will be the name of the measurement.
	Name string

	// Which tags to inherit from the top-level config.
	InheritTags []string

	// Adds each row's table index as a tag.
	IndexAsTag bool

	// Fields is the tags and values to look up.
	Fields []Field `toml:"field"`

	// OID for automatic field population.
	// If provided, init() will populate Fields with all the table columns of the
	// given OID.
	Oid string

	initialized bool
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

	initialized bool
}
