package applications

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	applicationsdomain "flatty-budget/go-api/domains/applications"
)

// pgxPool is a minimal interface matching the Query and QueryRow methods of *pgxpool.Pool.
// It exists to enable unit testing with mock implementations.
type pgxPool interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PgxRepository struct {
	pool pgxPool
}

func NewPgxRepository(pool pgxPool) *PgxRepository {
	return &PgxRepository{pool: pool}
}

const selectColumns = `
	id, name, env, bundle_js, COALESCE(style_url, ''), remote_origin, proxy_base_path, base_path, created_at, updated_at
`

func scanApplication(row pgx.Row) (*applicationsdomain.Application, error) {
	var id int64
	var name, env, bundleJS, styleURL, remoteOrigin, proxyBasePath, basePath string
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&id, &name, &env, &bundleJS, &styleURL,
		&remoteOrigin, &proxyBasePath, &basePath, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	return applicationsdomain.NewApplication(
		id, name, env, bundleJS, styleURL,
		remoteOrigin, proxyBasePath, basePath, createdAt, updatedAt,
	), nil
}

func (r *PgxRepository) ListByEnv(ctx context.Context, env string) ([]*applicationsdomain.Application, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+selectColumns+`
		FROM applications
		WHERE env = $1
		ORDER BY name
	`, env)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var apps []*applicationsdomain.Application

	for rows.Next() {
		app, err := scanApplication(rows)
		if err != nil {
			return nil, err
		}

		apps = append(apps, app)
	}

	return apps, rows.Err()
}

func (r *PgxRepository) ListAll(ctx context.Context) ([]*applicationsdomain.Application, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+selectColumns+`
		FROM applications
		ORDER BY env, name
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var apps []*applicationsdomain.Application

	for rows.Next() {
		app, err := scanApplication(rows)
		if err != nil {
			return nil, err
		}

		apps = append(apps, app)
	}

	return apps, rows.Err()
}

func (r *PgxRepository) Create(ctx context.Context, input *applicationsdomain.ApplicationInput) (*applicationsdomain.Application, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO applications (name, env, bundle_js, style_url, remote_origin, proxy_base_path, base_path)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING `+selectColumns,
		input.Name(), input.Env(), input.BundleJS(), input.StyleURL(),
		input.RemoteOrigin(), input.ProxyBasePath(), input.BasePath(),
	)

	return scanApplication(row)
}

func (r *PgxRepository) Update(ctx context.Context, id int64, input *applicationsdomain.ApplicationInput) (*applicationsdomain.Application, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE applications
		SET
			name = $1,
			env = $2,
			bundle_js = $3,
			style_url = $4,
			remote_origin = $5,
			proxy_base_path = $6,
			base_path = $7
		WHERE id = $8
		RETURNING `+selectColumns,
		input.Name(), input.Env(), input.BundleJS(), input.StyleURL(),
		input.RemoteOrigin(), input.ProxyBasePath(), input.BasePath(), id,
	)

	app, err := scanApplication(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("application with id %d not found: %w", id, pgx.ErrNoRows)
		}

		return nil, err
	}

	return app, nil
}

func (r *PgxRepository) Delete(ctx context.Context, id int64) (int64, error) {
	var returningID int64

	err := r.pool.QueryRow(ctx, `
		DELETE FROM applications
		WHERE id = $1
		RETURNING id
	`, id).Scan(&returningID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, fmt.Errorf("application with id %d not found: %w", id, pgx.ErrNoRows)
		}

		return -1, err
	}

	return returningID, nil
}
