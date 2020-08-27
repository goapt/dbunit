package fixtures

import (
	"reflect"
	"testing"
	"time"
)

func Test_tryStrToDate(t *testing.T) {
	type args struct {
		loc *time.Location
		s   string
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name:    "time1",
			args:    args{time.Local, "2020-08-20"},
			want:    time.Date(2020, 8, 20, 0, 0, 0, 0, time.Local),
			wantErr: false,
		},
		{
			name:    "time2",
			args:    args{nil, "2020-08-20 12:12:12"},
			want:    time.Date(2020, 8, 20, 12, 12, 12, 0, time.Local),
			wantErr: false,
		},
		{
			name:    "time3",
			args:    args{loc, "2019-01-02T15:04:26+08:00"},
			want:    time.Date(2019, 1, 2, 15, 4, 26, 0, loc),
			wantErr: false,
		},
		{
			name:    "time4",
			args:    args{loc, "2019-0102T15:04:26+08:00"},
			want:    time.Time{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tryStrToDate(tt.args.loc, tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("tryStrToDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tryStrToDate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
