package domain

import (
	"context"
	"github.com/PaesslerAG/gval"
)

func CloneReadMapping(from, to *ReadMapping) {
	to.Id = from.Id
	to.Body = from.Body
	to.PreparedBody = from.PreparedBody
}

type PreparedBody map[string]gval.Evaluable
type ReadMapping struct {
	Id           string       `json:"id"`
	Body         Body         `json:"body"`
	PreparedBody PreparedBody `json:"-"`
}

type ReadMappingService interface {
	GetAll(ctx context.Context, result *[]ReadMapping) error
	Create(ctx context.Context, order *ReadMapping) error
	GetById(ctx context.Context, id string, result *ReadMapping) error
	DeleteById(ctx context.Context, id string) error
}
