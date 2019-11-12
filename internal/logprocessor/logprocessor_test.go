package logprocessor

import (
	"flag"
	"reflect"
	"testing"

	"github.com/shipt/specter/internal/maxmind"
)

var db = flag.String("db", "./GeoList2-City.mmdb", "path to maxmind db")

func TestNew(t *testing.T) {
	type args struct {
		maxMind maxmind.MaxMind
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.maxMind); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Write(t *testing.T) {
	type fields struct {
		maxMind maxmind.MaxMind
		qchan   chan struct{}
		rchan   chan []byte
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				maxMind: tt.fields.maxMind,
				qchan:   tt.fields.qchan,
				rchan:   tt.fields.rchan,
			}
			c.Write(tt.args.b)
		})
	}
}

func TestClient_parseResponse(t *testing.T) {
	type fields struct {
		maxMind maxmind.MaxMind
		qchan   chan struct{}
		rchan   chan []byte
	}
	type args struct {
		body []byte
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
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    connection
		wantErr bool
	}{
		{
			name: "Works with valid data",
			args: args{
				[]byte(`{"src_ip":"54.86.88.147","dst_ip":"146.148.55.211","http_status":"200","http_x_forwarded_for":"54.86.88.147"}`),
			},
			want:    connection{},
			wantErr: false,
		},
		{
			name: "Breaks with invalid data",
			args: args{
				[]byte(`{"src_ip":"54.86.88.147","dst_ip":"","http_status":"200","http_x_forwarded_for":"54.86.88.147"}`),
			},
			want:    connection{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				maxMind: tt.fields.maxMind,
				qchan:   tt.fields.qchan,
				rchan:   tt.fields.rchan,
			}
			got, err := c.parseResponse(tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.parseResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.parseResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Ingest(t *testing.T) {
	type fields struct {
		maxMind maxmind.MaxMind
		qchan   chan struct{}
		rchan   chan []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   <-chan []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				maxMind: tt.fields.maxMind,
				qchan:   tt.fields.qchan,
				rchan:   tt.fields.rchan,
			}
			if got := c.Ingest(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Ingest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Stop(t *testing.T) {
	type fields struct {
		maxMind maxmind.MaxMind
		qchan   chan struct{}
		rchan   chan []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   chan struct{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				maxMind: tt.fields.maxMind,
				qchan:   tt.fields.qchan,
				rchan:   tt.fields.rchan,
			}
			if got := c.Stop(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Stop() = %v, want %v", got, tt.want)
			}
		})
	}
}
