package httpserver

import (
	"context"
	"vc/internal/apigw/apiv1"
	apiv1_status "vc/internal/gen/status/apiv1.status"

	"go.opentelemetry.io/otel/codes"

	"github.com/gin-gonic/gin"
)

func (s *Service) endpointUpload(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.tp.Start(ctx, "httpserver:endpointUpload")
	defer span.End()

	request := &apiv1.UploadRequest{}
	if err := s.bindRequest(ctx, c, request); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if s.kafkaMessageProducer != nil {
		err := s.kafkaMessageProducer.Upload(request)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		return nil, nil
	}

	err := s.apiv1.Upload(ctx, request)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return nil, nil
}

func (s *Service) endpointNotification(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.tp.Start(ctx, "httpserver:endpointNotification")
	defer span.End()

	request := &apiv1.NotificationRequest{}
	if err := s.bindRequest(ctx, c, request); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	reply, err := s.apiv1.Notification(ctx, request)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return reply, nil
}

func (s *Service) endpointGetDocument(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.tp.Start(ctx, "httpserver:endpointGetDocument")
	defer span.End()

	request := &apiv1.GetDocumentRequest{}
	if err := s.bindRequest(ctx, c, request); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	reply, err := s.apiv1.GetDocument(ctx, request)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return reply, nil
}

func (s *Service) endpointRevokeDocument(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.tp.Start(ctx, "httpserver:endpointRevokeDocument")
	defer span.End()

	request := &apiv1.RevokeDocumentRequest{}
	if err := s.bindRequest(ctx, c, request); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	if err := s.apiv1.RevokeDocument(ctx, request); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return nil, nil
}

func (s *Service) endpointDeleteDocument(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.tp.Start(ctx, "httpserver:endpointDeleteDocument")
	defer span.End()

	request := &apiv1.DeleteDocumentRequest{}
	if err := s.bindRequest(ctx, c, request); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	err := s.apiv1.DeleteDocument(ctx, request)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return nil, nil
}

func (s *Service) endpointGetDocumentCollectID(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.tp.Start(ctx, "httpserver:endpointGetDocumentAttestation")
	defer span.End()

	request := &apiv1.GetDocumentCollectIDRequest{}
	if err := s.bindRequest(ctx, c, request); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	reply, err := s.apiv1.GetDocumentCollectID(ctx, request)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return reply, nil
}

func (s *Service) endpointIDMapping(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.tp.Start(ctx, "httpserver:endpointIDMapping")
	defer span.End()

	request := &apiv1.IDMappingRequest{}
	if err := s.bindRequest(ctx, c, request); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	reply, err := s.apiv1.IDMapping(ctx, request)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return reply, nil
}

func (s *Service) endpointPortal(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.tp.Start(ctx, "httpserver:endpointPortal")
	defer span.End()

	request := &apiv1.PortalRequest{}
	if err := s.bindRequest(ctx, c, request); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	reply, err := s.apiv1.Portal(ctx, request)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return reply, nil
}

func (s *Service) endpointHealth(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.tp.Start(ctx, "httpserver:endpointHealth")
	defer span.End()

	request := &apiv1_status.StatusRequest{}
	reply, err := s.apiv1.Health(ctx, request)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (s *Service) endpointCredential(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.tp.Start(ctx, "httpserver:endpointCredential")
	defer span.End()

	request := &apiv1.CredentialRequest{}
	if err := s.bindRequest(ctx, c, request); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	reply, err := s.apiv1.Credential(ctx, request)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return reply, nil
}
