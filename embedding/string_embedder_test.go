package embedding

import (
	"context"
	"reflect"
	"testing"
)

func TestStringEmbedder_Embed(t *testing.T) {
	type args struct {
		ctx  context.Context
		text string
	}
	tests := []struct {
		name    string
		s       StringEmbedder
		args    args
		want    []float32
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			s:    StringEmbedder{},
			args: args{
				ctx:  context.Background(),
				text: "hello world",
			},
			want:    []float32{104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StringEmbedder{}
			got, err := s.Embed(tt.args.ctx, tt.args.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringEmbedder.Embed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringEmbedder.Embed() = %v, want %v", got, tt.want)
			}
		})
	}
}
