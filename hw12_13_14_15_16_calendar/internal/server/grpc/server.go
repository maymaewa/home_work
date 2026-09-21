package grpc

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	event "github.com/maymaewa/home_work/hw12_13_14_15_calendar/api"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Config struct {
	Host string
	Port int
}

type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) (storage.Event, error)
	UpdateEvent(ctx context.Context, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error

	ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForWeek(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForMonth(ctx context.Context, date time.Time) ([]storage.Event, error)
}

type Logger interface {
	InfoContext(msg string, args ...any)
	Error(msg string)
}

type Server struct {
	app    Application
	logger Logger
	server *grpc.Server
	config Config
}

type eventService struct {
	event.UnimplementedEventServiceServer

	app Application
}

func loggingInterceptor(logger Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		response, err := handler(ctx, req)

		duration := time.Since(start)

		if err != nil {
			logger.Error(
				"gRPC request failed: " + info.FullMethod +
					", duration: " + duration.String() +
					", error: " + err.Error(),
			)

			return response, err
		}

		logger.InfoContext(
			"gRPC request completed",
			"method", info.FullMethod,
			"duration", duration,
		)

		return response, nil
	}
}

func NewServer(logger Logger, app Application, config Config) *Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(logger)),
	)

	event.RegisterEventServiceServer(grpcServer, &eventService{
		app: app,
	})

	return &Server{
		app:    app,
		logger: logger,
		server: grpcServer,
		config: config,
	}
}

func (s *Server) Start(ctx context.Context) error {
	address := net.JoinHostPort(
		s.config.Host,
		strconv.Itoa(s.config.Port),
	)

	var listenConfig net.ListenConfig

	listener, err := listenConfig.Listen(ctx, "tcp", address)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		s.server.GracefulStop()
	}()

	err = s.server.Serve(listener)
	if errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}

	return err
}

func (s *Server) Stop(_ context.Context) error {
	s.server.GracefulStop()

	return nil
}

func (s *eventService) CreateEvent(
	ctx context.Context,
	req *event.CreateEventRequest,
) (*event.Event, error) {
	if req == nil || req.GetEvent() == nil {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}

	model, err := eventFromProto(req.GetEvent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	createdEvent, err := s.app.CreateEvent(ctx, model)
	if err != nil {
		return nil, mapStorageError(err)
	}

	return eventToProto(createdEvent), nil
}

func (s *eventService) UpdateEvent(
	ctx context.Context,
	req *event.UpdateEventRequest,
) (*event.Event, error) {
	if req == nil || req.GetEvent() == nil {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}

	model, err := eventFromProto(req.GetEvent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	model.ID = req.GetId()

	if err := s.app.UpdateEvent(ctx, model); err != nil {
		return nil, mapStorageError(err)
	}

	return eventToProto(model), nil
}

func (s *eventService) DeleteEvent(
	ctx context.Context,
	req *event.DeleteEventRequest,
) (*emptypb.Empty, error) {
	if req == nil || req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := s.app.DeleteEvent(ctx, req.GetId()); err != nil {
		return nil, mapStorageError(err)
	}

	return &emptypb.Empty{}, nil
}

func (s *eventService) ListEventsForDay(
	ctx context.Context,
	req *event.ListEventsRequest,
) (*event.ListEventsResponse, error) {
	return s.listEvents(ctx, req, s.app.ListEventsForDay)
}

func (s *eventService) ListEventsForWeek(
	ctx context.Context,
	req *event.ListEventsRequest,
) (*event.ListEventsResponse, error) {
	return s.listEvents(ctx, req, s.app.ListEventsForWeek)
}

func (s *eventService) ListEventsForMonth(
	ctx context.Context,
	req *event.ListEventsRequest,
) (*event.ListEventsResponse, error) {
	return s.listEvents(ctx, req, s.app.ListEventsForMonth)
}

type listEventsFunc func(context.Context, time.Time) ([]storage.Event, error)

func (s *eventService) listEvents(
	ctx context.Context,
	req *event.ListEventsRequest,
	listFunc listEventsFunc,
) (*event.ListEventsResponse, error) {
	if req == nil || req.GetDate() == nil {
		return nil, status.Error(codes.InvalidArgument, "date is required")
	}

	if err := req.GetDate().CheckValid(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid date")
	}

	events, err := listFunc(ctx, req.GetDate().AsTime())
	if err != nil {
		return nil, mapStorageError(err)
	}

	response := &event.ListEventsResponse{
		Events: make([]*event.Event, 0, len(events)),
	}

	for _, item := range events {
		response.Events = append(response.Events, eventToProto(item))
	}

	return response, nil
}

func eventFromProto(src *event.Event) (storage.Event, error) {
	if src == nil {
		return storage.Event{}, errors.New("event is required")
	}

	if src.GetStartAt() == nil {
		return storage.Event{}, errors.New("start_at is required")
	}

	if err := src.GetStartAt().CheckValid(); err != nil {
		return storage.Event{}, errors.New("invalid start_at")
	}

	if src.GetEndAt() == nil {
		return storage.Event{}, errors.New("end_at is required")
	}

	if err := src.GetEndAt().CheckValid(); err != nil {
		return storage.Event{}, errors.New("invalid end_at")
	}

	result := storage.Event{
		ID:          src.GetId(),
		Title:       src.GetTitle(),
		StartAt:     src.GetStartAt().AsTime(),
		EndAt:       src.GetEndAt().AsTime(),
		Description: src.GetDescription(),
		UserID:      src.GetUserId(),
	}

	if src.GetNotifyBefore() != nil {
		if err := src.GetNotifyBefore().CheckValid(); err != nil {
			return storage.Event{}, errors.New("invalid notify_before")
		}

		duration := src.GetNotifyBefore().AsDuration()
		result.NotifyBefore = &duration
	}

	return result, nil
}

func eventToProto(src storage.Event) *event.Event {
	result := &event.Event{
		Id:          src.ID,
		Title:       src.Title,
		StartAt:     timestamppb.New(src.StartAt),
		EndAt:       timestamppb.New(src.EndAt),
		Description: src.Description,
		UserId:      src.UserID,
	}

	if src.NotifyBefore != nil {
		result.NotifyBefore = durationToProto(*src.NotifyBefore)
	}

	return result
}

func durationToProto(value time.Duration) *durationpb.Duration {
	return durationpb.New(value)
}

func mapStorageError(err error) error {
	switch {
	case errors.Is(err, storage.ErrEventNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, storage.ErrEventAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, storage.ErrDateBusy):
		return status.Error(codes.FailedPrecondition, err.Error())

	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())

	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())

	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
