package infrastructure

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ctx(t *testing.T) {
	type args struct {
		logerVal   string
		txVal      string
		requestVal string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Case #1. Get value",
			args: args{
				logerVal: "logerVal",
			},
			wantErr: false,
		},
		{
			name: "Case #2. Get value",
			args: args{
				logerVal:   "logerVal",
				requestVal: "requestVal",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.args.logerVal != "" {
				ctx = context.WithValue(ctx, CtxKeyLogger{}, tt.args.logerVal)
			}
			if tt.args.txVal != "" {
				ctx = context.WithValue(ctx, CtxKeyTransaction{}, tt.args.txVal)
			}
			if tt.args.requestVal != "" {
				ctx = context.WithValue(ctx, CtxKeyRequestID{}, tt.args.requestVal)
			}
			ctxValue := ctx.Value(CtxKeyLogger{})
			require.NotEmpty(t, ctxValue)
			assert.Equal(t, ctxValue, tt.args.logerVal)
		})
	}
}
