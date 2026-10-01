package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	event "github.com/maymaewa/home_work/hw12_13_14_15_calendar/api"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/app"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/logger"
	memorystorage "github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	testHost        = "127.0.0.1"
	testHostPort    = testHost + ":0"
	testUserID      = "user-1"
	testCRUDEventID = "crud-test-event"
)

func TestCreateEvent(t *testing.T) {
	storage := memorystorage.New()
	logg := logger.New("error")
	calendar := app.New(logg, storage)

	server := NewServer(
		logg,
		calendar,
		Config{
			Host: testHost,
			Port: 0,
		},
	)

	var listenConfig net.ListenConfig

	listener, err := listenConfig.Listen(context.Background(), "tcp", testHostPort)
	require.NoError(t, err)

	go func() {
		_ = server.server.Serve(listener)
	}()

	t.Cleanup(func() {
		server.server.Stop()
		_ = listener.Close()
	})

	conn, err := grpc.NewClient(
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = conn.Close()
	})

	client := event.NewEventServiceClient(conn)

	startAt := time.Date(2026, 9, 22, 15, 0, 0, 0, time.UTC)
	endAt := startAt.Add(time.Hour)

	response, err := client.CreateEvent(
		context.Background(),
		&event.CreateEventRequest{
			Event: &event.Event{
				Id:          "test-event-id",
				Title:       "gRPC test event",
				StartAt:     timestamppb.New(startAt),
				EndAt:       timestamppb.New(endAt),
				Description: "created through gRPC",
				UserId:      testUserID,
				NotifyBefore: durationpb.New(
					30 * time.Minute,
				),
			},
		},
	)

	require.NoError(t, err)
	require.NotNil(t, response)

	require.NotEmpty(t, response.Id)
	require.Equal(t, "gRPC test event", response.Title)
	require.Equal(t, testUserID, response.UserId)
	require.Equal(t, "created through gRPC", response.Description)
	require.Equal(t, 30*time.Minute, response.NotifyBefore.AsDuration())
	require.True(t, response.StartAt.AsTime().Equal(startAt))
	require.True(t, response.EndAt.AsTime().Equal(endAt))
}

func TestUpdateListDeleteEvent(t *testing.T) {
	storage := memorystorage.New()
	logg := logger.New("error")
	calendar := app.New(logg, storage)

	server := NewServer(
		logg,
		calendar,
		Config{
			Host: testHost,
			Port: 0,
		},
	)

	var listenConfig net.ListenConfig

	listener, err := listenConfig.Listen(context.Background(), "tcp", testHostPort)
	require.NoError(t, err)

	go func() {
		_ = server.server.Serve(listener)
	}()

	t.Cleanup(func() {
		server.server.Stop()
		_ = listener.Close()
	})

	conn, err := grpc.NewClient(
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = conn.Close()
	})

	client := event.NewEventServiceClient(conn)

	startAt := time.Date(2026, 9, 23, 15, 0, 0, 0, time.UTC)
	endAt := startAt.Add(time.Hour)

	created, err := client.CreateEvent(
		context.Background(),
		&event.CreateEventRequest{
			Event: &event.Event{
				Id:      testCRUDEventID,
				Title:   "Original title",
				StartAt: timestamppb.New(startAt),
				EndAt:   timestamppb.New(endAt),
				UserId:  testUserID,
			},
		},
	)

	require.NoError(t, err)
	require.Equal(t, testCRUDEventID, created.Id)

	updated, err := client.UpdateEvent(
		context.Background(),
		&event.UpdateEventRequest{
			Id: testCRUDEventID,
			Event: &event.Event{
				Id:          testCRUDEventID,
				Title:       "Updated title",
				StartAt:     timestamppb.New(startAt),
				EndAt:       timestamppb.New(endAt),
				Description: "updated description",
				UserId:      testUserID,
				NotifyBefore: durationpb.New(
					15 * time.Minute,
				),
			},
		},
	)

	require.NoError(t, err)
	require.Equal(t, testCRUDEventID, updated.Id)
	require.Equal(t, "Updated title", updated.Title)
	require.Equal(t, "updated description", updated.Description)
	require.Equal(t, 15*time.Minute, updated.NotifyBefore.AsDuration())

	list, err := client.ListEventsForDay(
		context.Background(),
		&event.ListEventsRequest{
			Date:   timestamppb.New(startAt),
			UserId: testUserID,
		},
	)

	require.NoError(t, err)
	require.Len(t, list.Events, 1)
	require.Equal(t, testCRUDEventID, list.Events[0].Id)
	require.Equal(t, "Updated title", list.Events[0].Title)

	_, err = client.DeleteEvent(
		context.Background(),
		&event.DeleteEventRequest{
			Id: testCRUDEventID,
		},
	)

	require.NoError(t, err)

	list, err = client.ListEventsForDay(
		context.Background(),
		&event.ListEventsRequest{
			Date:   timestamppb.New(startAt),
			UserId: testUserID,
		},
	)

	require.NoError(t, err)
	require.Empty(t, list.Events)
}

func TestEventErrors(t *testing.T) {
	storage := memorystorage.New()
	logg := logger.New("error")
	calendar := app.New(logg, storage)

	server := NewServer(
		logg,
		calendar,
		Config{
			Host: testHost,
			Port: 0,
		},
	)

	var listenConfig net.ListenConfig

	listener, err := listenConfig.Listen(context.Background(), "tcp", testHostPort)
	require.NoError(t, err)

	go func() {
		_ = server.server.Serve(listener)
	}()

	t.Cleanup(func() {
		server.server.Stop()
		_ = listener.Close()
	})

	conn, err := grpc.NewClient(
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = conn.Close()
	})

	client := event.NewEventServiceClient(conn)

	startAt := time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC)
	endAt := startAt.Add(time.Hour)

	request := &event.CreateEventRequest{
		Event: &event.Event{
			Id:      "duplicate-event",
			Title:   "Test event",
			StartAt: timestamppb.New(startAt),
			EndAt:   timestamppb.New(endAt),
			UserId:  testUserID,
		},
	}

	_, err = client.CreateEvent(context.Background(), request)
	require.NoError(t, err)

	_, err = client.CreateEvent(context.Background(), request)
	require.Error(t, err)
	require.Equal(t, codes.AlreadyExists, status.Code(err))

	_, err = client.DeleteEvent(
		context.Background(),
		&event.DeleteEventRequest{
			Id: "does-not-exist",
		},
	)

	require.Error(t, err)
	require.Equal(t, codes.NotFound, status.Code(err))
}
