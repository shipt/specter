package dataServer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/shipt/specter/cmd"

	"github.com/namsral/flag"
	"github.com/pkg/errors"

	externalip "github.com/GlenDC/go-external-ip"
	"github.com/hpcloud/tail"

	"github.com/satyrius/gonx"
	log "github.com/sirupsen/logrus"
)

var previousOffset int64

type msg struct {
	SrcIP      string `json:"src_ip"`
	DstIP      string `json:"dst_ip"`
	HTTPStatus string `json:"http_status"`
}

type tailReader struct {
	*tail.Tail
	cur bytes.Buffer
}

type ngninxLogReader interface {
	Read() (*gonx.Entry, error)
}

var conf string
var format string
var logFile string
var server string

func init() {
	flag.StringVar(&conf, "conf", "", "Nginx config file (e.g. /etc/nginx/nginx.conf)")
	flag.StringVar(&format, "format", "main", "Nginx log_format name")
	flag.StringVar(&logFile, "log", "", "The location of the access.log file. Reads from STDIN if no value is set")
	flag.StringVar(&server, "server", "http://localhost:1323", "The Specter webserver's server IP:Port")
}

func (t *tailReader) Read(b []byte) (int, error) {
	if t.cur.Len() == 0 {
		text := (<-t.Lines).Text
		t.cur.WriteString(text)
		t.cur.WriteByte('\n')
	}

	n, err := t.cur.Read(b)
	if err == io.EOF {
		return n, nil
	}

	return n, err
}

func getExternalIP() (net.IP, error) {
	exip := externalip.DefaultConsensus(nil, nil)
	ip, err := exip.ExternalIP()
	if err != nil {
		return ip, errors.Wrap(err, "error getting external IP")
	}
	log.Info("Server IP:", ip.String())
	return ip, nil
}

func tailFile(logFile string) (*tail.Tail, error) {
	log.Debug("Starting file tail")
	tail, err := tail.TailFile(logFile, tail.Config{Logger: tail.DiscardingLogger, Follow: true, ReOpen: true, Location: &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}})
	log.Debugf("%+v\n", tail)
	return tail, errors.Wrap(err, "Error tailing file")
}

func logReader(tail *tail.Tail) *tailReader {
	log.Debug("Starting log reader")
	return &tailReader{Tail: tail}
}

func sendMessage(url string, mBytes []byte) error {
	_, err := http.Post(url, "application/json", bytes.NewReader(mBytes))
	log.Debug("sending data to webserver")
	return errors.Wrap(err, "error sending message")
}

func processLog(reader ngninxLogReader, ip net.IP) (msg, error) {
	rec, err := reader.Read()
	log.Debugf("%+v", rec)
	log.Debug("Reading log entry")
	if err == io.EOF {
		log.Debug("reached EOF of log file")
		return msg{}, nil
	}
	if err != nil {
		return msg{}, errors.Wrap(err, "error reading the log file")
	}
	log.Debug("log entry not EOF")
	// Process the record...
	ra, err := rec.Field("remote_addr")
	log.Debug("processing the log entry")
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Warn("error getting the remote address from the access.log")
		return msg{}, nil
	}
	log.Debugf("ra: %s", ra)
	s, err := rec.Field("status")
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Warn("error getting the status from the access.log")
		return msg{}, nil
	}
	log.Debugf("status: %s", s)
	log.Debug("returning processed message")
	return msg{SrcIP: ra, DstIP: ip.String(), HTTPStatus: s}, nil
}

// Start starts and runs the data server
func Start() {
	flag.Parse()
	if !cmd.IsFlagPassed("conf") {
		log.Fatal(`you did not set a Nginx Config File.`)
	}
	if !cmd.IsFlagPassed("format") {
		log.Warn(`you did not set a Nginx log_format name, using "main"`)
	}
	if !cmd.IsFlagPassed("log") {
		log.Warn(`you did not set a log location, using STDIN"`)
	}
	if !cmd.IsFlagPassed("server") {
		log.Warn(`you did not set a server to send data to, using localhost:1323`)
	}

	log.Info("Starting Dataserver")
	log.Debugf("conf flag is set to: %s", conf)
	log.Debugf("format flag is set to: %s", format)
	log.Debugf("log flag is set to: %s", logFile)
	log.Debugf("server flag is set to: %s", server)

	url := fmt.Sprintf("%s/logs", server)
	ip, err := getExternalIP()
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("error calling getExteranlIp")
	}

	tail, err := tailFile(logFile)
	log.Debug("Starting tailFile")
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("error calling tailFile")
	}

	cf, err := os.Open(conf)
	log.Debug("Opening config file")
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("error opening nginx config file")
	}
	defer cf.Close()

	reader, err := gonx.NewNginxReader(logReader(tail), cf, format)
	log.Debug("Reading config file")
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("error creating reader the nginx config file")
	}

	for {
		m, err := processLog(reader, ip)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
			}).Fatal("error processing log file")
		}
		if (msg{} == m) {
			continue
		}

		mBytes, err := json.Marshal(m)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
			}).Warn("error serializing message")
		}
		err = sendMessage(url, mBytes)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
			}).Warn("error posting to specter webserver")
		}
	}

}
