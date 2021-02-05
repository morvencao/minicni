package nettool

import (
	"reflect"
	"testing"
)

func TestGetAllIPs(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:    "valid cidr input",
			input:   "192.168.0.0/30",
			want:    []string{"192.168.0.1/30", "192.168.0.2/30"},
			wantErr: false,
		},
		{
			name:    "invalid cidr input",
			input:   "192.168.0/30",
			want:    nil,
			wantErr: true,
		},
		{
			name:  "valid cidr input with more ips",
			input: "62.76.47.12/28",
			want: []string{"62.76.47.1/28", "62.76.47.2/28", "62.76.47.3/28", "62.76.47.4/28",
				"62.76.47.5/28", "62.76.47.6/28", "62.76.47.7/28", "62.76.47.8/28", "62.76.47.9/28",
				"62.76.47.10/28", "62.76.47.11/28", "62.76.47.12/28", "62.76.47.13/28", "62.76.47.14/28"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allips, err := GetAllIPs(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllIPs error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(allips, tt.want) {
				t.Errorf("wanted:\n%v\ngot:\n%v", tt.want, allips)
			}
		})
	}
}
