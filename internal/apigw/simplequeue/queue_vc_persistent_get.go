package simplequeue

import (
	"context"
	"encoding/json"
	"vc/pkg/logger"

	retask "github.com/masv3971/goretask"
	"go.opentelemetry.io/otel/codes"
)

// VCPersistentGet holds the ladok delete signed queue
type VCPersistentGet struct {
	service *Service
	log     *logger.Log
	*retask.Queue
}

// NewVCPersistentGet creates a new queue for getting documents from the persistent queue
func NewVCPersistentGet(ctx context.Context, service *Service, queueName string, log *logger.Log) (*VCPersistentGet, error) {
	vcPersistentGet := &VCPersistentGet{
		service: service,
		log:     log,
	}

	vcPersistentGet.Queue = vcPersistentGet.service.queueClient.NewQueue(ctx, queueName)

	vcPersistentGet.log.Info("Started")

	return vcPersistentGet, nil
}

// Enqueue publishes a document to the queue
func (s *VCPersistentGet) Enqueue(ctx context.Context, message any) (*retask.Job, error) {
	s.log.Info("Enqueue")
	ctx, span := s.service.tp.Start(ctx, "simplequeue:VCPersistentGet:Enqueue")
	defer span.End()

	data, err := json.Marshal(message)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return s.Queue.Enqueue(ctx, data)
}

// Dequeue dequeues a document from the queue
func (s *VCPersistentGet) Dequeue(ctx context.Context) error {
	ctx, span := s.service.tp.Start(ctx, "simplequeue:VCPersistentGet:Dequeue")
	defer span.End()
	return nil
}

// Wait waits for the next message
func (s *VCPersistentGet) Wait(ctx context.Context) (*retask.Task, error) {
	ctx, span := s.service.tp.Start(ctx, "simplequeue:VCPersistentGet:Wait")
	defer span.End()

	task, err := s.Queue.Wait(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return task, nil
}
