package domain

import "context"

type Body map[string]interface{}

type TxFunc func(txCtx context.Context) error
type ExecTxFunc func(ctx context.Context, f TxFunc) error
