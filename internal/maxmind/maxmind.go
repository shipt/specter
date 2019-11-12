package maxmind

import (
	"fmt"
	"net"

	"github.com/namsral/flag"

	maxminddb "github.com/oschwald/maxminddb-golang"
)

// Record contains the information we want out of maxMind
type Record struct {
	Location struct {
		Latitude  float32 `maxminddb:"latitude"`
		Longitude float32 `maxminddb:"longitude"`
	} `maxminddb:"location"`
}

// MaxMind contains a pointer to a reader
type MaxMind struct {
	db *maxminddb.Reader
}

// New creates a new maxMind reader.
func New(filename string) (MaxMind, error) {
	mmdb := flag.Lookup("db").Value.(flag.Getter).Get().(string)
	println("Maxmind DB found at ", mmdb)
	db, err := maxminddb.Open(mmdb)
	if err != nil {
		return MaxMind{}, fmt.Errorf("Failed opening the MaxMind db: %+v", err)
	}
	return MaxMind{db}, nil
}

// LookupIP returns geo information about an IP
func (mm MaxMind) LookupIP(ipStr string) (Record, error) {
	ip := net.ParseIP(ipStr)
	var record Record
	err := mm.db.Lookup(ip, &record)
	return record, err
}

// Close closes the maxMind connection
func (mm MaxMind) Close() error {
	return mm.db.Close()
}
