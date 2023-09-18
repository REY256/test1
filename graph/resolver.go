package graph

import (
	"test1/internal/app/service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	s *service.Service
}

func New(s *service.Service) *Resolver {
	return &Resolver{s}
}
