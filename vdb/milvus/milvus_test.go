package milvus

import (
	"context"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		want    *VDB
		wantErr bool
	}{
		{
			name: "succ",
			cfg: Config{
				Address:    "9.134.44.65:19530",
				Collection: "test_table",
				Dimension:  1024,
			},
			want:    nil,
			wantErr: false,
		},
	}
	db, err := New(tests[0].cfg)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	b, err := db.client.HasCollection(context.TODO(), tests[0].cfg.Collection)
	fmt.Println("exist:", b)
	fmt.Println("err:", err)
}
