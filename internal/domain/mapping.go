package domain

import (
	"context"
	"github.com/PaesslerAG/gval"
)

type ReadMapping struct {
	Id           string                    `json:"id"`
	Body         Body                      `json:"body"`
	PreparedBody map[string]gval.Evaluable `json:"-"`
}

type ReadMappingService interface {
	GetAll(ctx context.Context) ([]ReadMapping, error)
	Create(ctx context.Context, order *ReadMapping) (*ReadMapping, error)
	GetById(ctx context.Context, id string) (*ReadMapping, error)
	DeleteById(ctx context.Context, id string) error
}
