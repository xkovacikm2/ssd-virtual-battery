package readings

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// KWh is an energy amount in hundredths of a kWh, matching numeric(10,2).
type KWh int64

func (k KWh) MarshalJSON() ([]byte, error) {
	sign := ""
	v := int64(k)
	if v < 0 {
		sign, v = "-", -v
	}
	frac := strconv.FormatInt(v%100, 10)
	if len(frac) == 1 {
		frac = "0" + frac
	}
	return []byte(sign + strconv.FormatInt(v/100, 10) + "." + frac), nil
}

func (k KWh) String() string {
	b, _ := k.MarshalJSON()
	return string(b)
}

type Reading struct {
	Date     time.Time
	Exported KWh
	Imported KWh
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Upsert(ctx context.Context, date time.Time, exported, imported KWh) error {
	if exported < 0 || imported < 0 {
		return errors.New("readings must not be negative")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO virtual_battery_readings (date, exported_to_grid, imported_from_grid, created_at, updated_at)
		VALUES ($1, $2::numeric / 100, $3::numeric / 100, now(), now())
		ON CONFLICT (date) DO UPDATE
		SET exported_to_grid = EXCLUDED.exported_to_grid,
		    imported_from_grid = EXCLUDED.imported_from_grid,
		    updated_at = now()`,
		date, int64(exported), int64(imported))
	return err
}

// Range returns readings with from <= date <= to, oldest first.
func (s *Store) Range(ctx context.Context, from, to time.Time) ([]Reading, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT date, (COALESCE(exported_to_grid, 0) * 100)::bigint, (COALESCE(imported_from_grid, 0) * 100)::bigint
		FROM virtual_battery_readings
		WHERE date BETWEEN $1 AND $2
		ORDER BY date`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Reading{}
	for rows.Next() {
		var r Reading
		var e, i int64
		if err := rows.Scan(&r.Date, &e, &i); err != nil {
			return nil, err
		}
		r.Exported, r.Imported = KWh(e), KWh(i)
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) MaxDate(ctx context.Context) (time.Time, bool, error) {
	var t *time.Time
	if err := s.pool.QueryRow(ctx, `SELECT max(date) FROM virtual_battery_readings`).Scan(&t); err != nil {
		return time.Time{}, false, err
	}
	if t == nil {
		return time.Time{}, false, nil
	}
	return *t, true, nil
}

func (s *Store) exists(ctx context.Context, where string, args ...any) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM virtual_battery_readings WHERE `+where+`)`, args...).Scan(&ok)
	return ok, err
}

func (s *Store) Any(ctx context.Context) (bool, error) {
	return s.exists(ctx, "true")
}

func (s *Store) AnyOnOrAfter(ctx context.Context, date time.Time) (bool, error) {
	return s.exists(ctx, "date >= $1", date)
}

func (s *Store) AnyBefore(ctx context.Context, date time.Time) (bool, error) {
	return s.exists(ctx, "date < $1", date)
}
