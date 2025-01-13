package chunk

import (
	"reflect"
	"testing"
)

func TestBase_mergeSplits(t *testing.T) {
	type fields struct {
		chunkSize   int
		overlapSize int
	}
	type args struct {
		splits    []string
		separator string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			fields: fields{
				chunkSize:   21,
				overlapSize: 5,
			},
			args: args{
				splits:    []string{"1234567890", "1234567890", "1234567890", "1234567890"},
				separator: ",",
			},
			want: []string{"1234567890,1234567890", "1234567890,1234567890"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Base{
				chunkSize:   tt.fields.chunkSize,
				overlapSize: tt.fields.overlapSize,
			}
			if got := b.mergeSplits(tt.args.splits, tt.args.separator); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Base.mergeSplits() = %v, want %v", got, tt.want)
			}
		})
	}
}
