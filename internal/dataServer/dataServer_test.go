package dataServer

import (
	"io"
	"net"
	"reflect"
	"testing"

	"github.com/satyrius/gonx"
)

type fakeLogReader struct {
	entry *gonx.Entry
}

func (flr fakeLogReader) Read() (*gonx.Entry, error) {
	return flr.entry, nil
}

type eofLogReader struct {
}

func (elr eofLogReader) Read() (*gonx.Entry, error) {
	return gonx.NewEmptyEntry(), io.EOF
}

type errLogReader struct {
}

func (erlr errLogReader) Read() (*gonx.Entry, error) {
	return gonx.NewEmptyEntry(), io.ErrUnexpectedEOF
}

func Test_processLog(t *testing.T) {
	type args struct {
		reader ngninxLogReader
		ip     net.IP
	}
	tests := []struct {
		name    string
		args    args
		want    msg
		wantErr bool
	}{
		{
			name: "Works with valid data",
			args: args{
				reader: fakeLogReader{
					entry: gonx.NewEntry(map[string]string{
						"remote_addr": "24.172.192.104",
						"status":      "200",
					})},
				ip: net.IP{},
			},
			want: msg{
				SrcIP:      "24.172.192.104",
				DstIP:      "<nil>",
				HTTPStatus: "200"},
			wantErr: false,
		},
		{
			name: "Returns msg{} with missing remote_addr",
			args: args{
				reader: fakeLogReader{
					entry: gonx.NewEntry(map[string]string{
						"status": "200",
					})},
				ip: net.IP{},
			},
			want:    msg{},
			wantErr: false,
		},
		{
			name: "Returns msg{} with missing status",
			args: args{
				reader: fakeLogReader{
					entry: gonx.NewEntry(map[string]string{
						"remote_addr": "24.172.192.104",
					})},
				ip: net.IP{},
			},
			want:    msg{},
			wantErr: false,
		},
		{
			name: "Doesnt error on EOF",
			args: args{
				reader: eofLogReader{},
				ip:     net.IP{},
			},
			want:    msg{},
			wantErr: false,
		},
		{
			name: "Errors on unexpected error",
			args: args{
				reader: errLogReader{},
				ip:     net.IP{},
			},
			want:    msg{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processLog(tt.args.reader, tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("processLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processLog() = %v, want %v", got, tt.want)
			}
		})
	}
}
