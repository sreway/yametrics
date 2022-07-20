package collector

import "testing"

func TestParseCounter(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    Counter
		wantErr bool
	}{
		{
			name: "success purse counter",
			args: args{
				s: "20",
			},
			want:    Counter(20),
			wantErr: false,
		},

		{
			name: "invalid purse counter",
			args: args{
				s: "none",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCounter(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCounter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseCounter() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseGause(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    Gauge
		wantErr bool
	}{
		{
			name: "success purse gauge",
			args: args{
				s: "7.7",
			},
			want:    Gauge(7.7),
			wantErr: false,
		},

		{
			name: "invalid purse gauge",
			args: args{
				s: "none",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGause(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGause() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseGause() got = %v, want %v", got, tt.want)
			}
		})
	}
}
