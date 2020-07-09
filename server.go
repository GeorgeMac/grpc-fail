package server

import (
	"context"
	"time"
)

type Service struct {
	Delay time.Duration
}

func NewServer(delay time.Duration) *Service {
	return &Service{delay}
}

func (s *Service) DoSlowThing(_ context.Context, _ *SlowThingRequest) (*SlowThingResponse, error) {
	time.Sleep(s.Delay)

	return &SlowThingResponse{}, nil
}
