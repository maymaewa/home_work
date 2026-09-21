package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/storage"
)

type Config struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

type Storage struct {
	db     *sql.DB
	config Config
}

func New(config Config) *Storage {
	return &Storage{
		config: config,
	}
}

func (s *Storage) Connect(ctx context.Context) error {
	dsn := (&url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(s.config.Username, s.config.Password),
		Host:   fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Path:   s.config.Database,
	}).String()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("ping database: %w", err)
	}

	s.db = db

	return nil
}

func (s *Storage) Close(_ context.Context) error {
	if s.db == nil {
		return nil
	}

	return s.db.Close()
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO events (
			id,
			title,
			start_at,
			end_at,
			description,
			user_id,
			notify_before
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		event.ID,
		event.Title,
		event.StartAt,
		event.EndAt,
		event.Description,
		event.UserID,
		notifyBeforeValue(event.NotifyBefore),
	)
	if err != nil {
		err = mapDBError(err)

		if errors.Is(err, storage.ErrEventAlreadyExists) {
			return err
		}

		return fmt.Errorf("create event: %w", err)
	}

	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, event storage.Event) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE events
		SET
			title = $2,
			start_at = $3,
			end_at = $4,
			description = $5,
			user_id = $6,
			notify_before = $7
		WHERE id = $1
	`,
		event.ID,
		event.Title,
		event.StartAt,
		event.EndAt,
		event.Description,
		event.UserID,
		notifyBeforeValue(event.NotifyBefore),
	)
	if err != nil {
		err = mapDBError(err)

		if errors.Is(err, storage.ErrDateBusy) {
			return err
		}

		return fmt.Errorf("update event: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}

	if rows == 0 {
		return storage.ErrEventNotFound
	}

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM events
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}

	if rows == 0 {
		return storage.ErrEventNotFound
	}

	return nil
}

func (s *Storage) listEvents(
	ctx context.Context,
	start time.Time,
	end time.Time,
) ([]storage.Event, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id,
			title,
			start_at,
			end_at,
			description,
			user_id,
			notify_before
		FROM events
		WHERE start_at < $2
		  AND end_at > $1
		ORDER BY start_at
	`, start, end)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	events := make([]storage.Event, 0)

	for rows.Next() {
		var event storage.Event
		var notifyBefore sql.NullInt64

		if err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.StartAt,
			&event.EndAt,
			&event.Description,
			&event.UserID,
			&notifyBefore,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}

		if notifyBefore.Valid {
			duration := time.Duration(notifyBefore.Int64)
			event.NotifyBefore = &duration
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}

	return events, nil
}

func (s *Storage) ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	start := startOfDay(date)

	return s.listEvents(
		ctx,
		start,
		start.AddDate(0, 0, 1),
	)
}

func (s *Storage) ListEventsForWeek(ctx context.Context, date time.Time) ([]storage.Event, error) {
	start := startOfWeek(date)

	return s.listEvents(
		ctx,
		start,
		start.AddDate(0, 0, 7),
	)
}

func (s *Storage) ListEventsForMonth(ctx context.Context, date time.Time) ([]storage.Event, error) {
	start := startOfMonth(date)

	return s.listEvents(
		ctx,
		start,
		start.AddDate(0, 1, 0),
	)
}

func startOfDay(t time.Time) time.Time {
	return time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		0, 0, 0, 0,
		t.Location(),
	)
}

func startOfWeek(t time.Time) time.Time {
	t = startOfDay(t)

	daysSinceMonday := (int(t.Weekday()) + 6) % 7

	return t.AddDate(0, 0, -daysSinceMonday)
}

func startOfMonth(t time.Time) time.Time {
	return time.Date(
		t.Year(),
		t.Month(),
		1,
		0, 0, 0, 0,
		t.Location(),
	)
}

func notifyBeforeValue(duration *time.Duration) any {
	if duration == nil {
		return nil
	}

	return int64(*duration)
}

func mapDBError(err error) error {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return storage.ErrEventAlreadyExists
		case "23P01":
			return storage.ErrDateBusy
		}
	}

	return err
}
