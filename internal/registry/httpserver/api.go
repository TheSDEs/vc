package httpserver

import (
	"context"
	"vc/internal/registry/apiv1"
	"vc/pkg/model"
)

// Apiv1 interface
type Apiv1 interface {
	Add(ctx context.Context, req *apiv1.AddRequest) (*apiv1.AddReply, error)
	Revoke(ctx context.Context, req *apiv1.RevokeRequest) (*apiv1.RevokeReply, error)
	Validate(ctx context.Context, req *apiv1.ValidateRequest) (*apiv1.ValidateReply, error)

	Status(ctx context.Context) (*model.Health, error)
}