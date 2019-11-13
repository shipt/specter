package logprocessor

import (
	"encoding/json"
	"net"
	"strings"

	"github.com/shipt/specter/internal/ttlcache"

	log "github.com/sirupsen/logrus"

	"github.com/shipt/specter/internal/maxmind"
)

// Client contains the information needed to communicate with maxMind
type Client struct {
	maxMind maxmind.MaxMind
	qchan   chan struct{}
	rchan   chan []byte
}

type params struct {
	SrcIP      string `json:"src_ip"`
	DstIP      string `json:"dst_ip"`
	HTTPStatus string `json:"http_status"`
	XFwdFor    string `json:"http_x_forwarded_for"`
}

type connection struct {
	DstIP      string
	SrcIP      string
	XFwdFor    string
	SrcLat     float32
	SrcLong    float32
	DstLat     float32
	DstLong    float32
	HTTPStatus string
}

var ipCache *ttlcache.TTLSet

func init() {
	ipCache = ttlcache.New(500)
	ipCache.Initalize()
}

// New creates a new client
func New(maxMind maxmind.MaxMind) *Client {
	return &Client{maxMind: maxMind, rchan: make(chan []byte)}
}

func (c *Client) Write(b []byte) {
	c.rchan <- b
}

func (c *Client) parseResponse(body []byte) ([]byte, error) {
	var rb params
	err := json.Unmarshal(body, &rb)
	if err != nil {
		return nil, err
	}

	conn := connection{}

	candidate := rb.XFwdFor
	if strings.Contains(rb.XFwdFor, ",") {
		parsedIPs := strings.Split(rb.XFwdFor, ",")
		candidate = parsedIPs[0]
	}

	conn.SrcIP = candidate
	validIP := net.ParseIP(candidate)
	if validIP == nil {
		conn.SrcIP = rb.SrcIP
	}

	conn.HTTPStatus = rb.HTTPStatus

	if ok := ipCache.Exist(conn.SrcIP); !ok {
		// we have seen this ip before.
		ipCache.Put(conn.SrcIP)
		srcRecord, err := c.maxMind.LookupIP(conn.SrcIP)
		if err != nil {
			return nil, err
		}
		conn.SrcLat = srcRecord.Location.Latitude
		conn.SrcLong = srcRecord.Location.Longitude

		conn.DstIP = rb.DstIP
		dstRecord, err := c.maxMind.LookupIP(conn.DstIP)
		if err != nil {
			return nil, err
		}
		conn.DstLat = dstRecord.Location.Latitude
		conn.DstLong = dstRecord.Location.Longitude

		cb, err := json.Marshal(conn)
		if err != nil {
			return nil, err
		}
		return cb, nil
	}
	return nil, nil
}

// Ingest ingests the logs channel
func (c *Client) Ingest() <-chan []byte {
	mchan := make(chan []byte)

	go func() {

		for {
			msg, err := c.parseResponse(<-c.rchan)
			if err != nil {
				log.WithFields(log.Fields{
					"Error": err,
				}).Fatal("error ingesting log channel")
			}
			select {
			case mchan <- msg:
			case <-c.qchan:
				return
			}
		}
	}()

	return mchan
}

// Stop handles closing maxMind
func (c *Client) Stop() chan struct{} {
	c.qchan <- struct{}{}
	if err := c.maxMind.Close(); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Panic("error stopping the log processor")
	}
	return c.qchan
}
