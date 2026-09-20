package app

import (
	"context"

	"backend/src/config"
	"backend/src/db"
	"backend/src/external"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Application struct {
	cfg     config.Config
	p       *pgxpool.Pool // dùng khi cần transaction: q.WithTx(tx)
	q       *db.Queries
	riot    *external.RiotClient
	ddragon *external.DDragonClient
}

func NewApplication(ctx context.Context, cfg config.Config) (*Application, error) {
	p, err := pgxpool.New(ctx, cfg.PostgresConnectionUrl)
	if err != nil {
		return nil, err
	}

	err = p.Ping(ctx)
	if err != nil {
		p.Close()
		return nil, err
	}

	return &Application{
		cfg:     cfg,
		p:       p,
		q:       db.New(p),
		riot:    external.NewRiotClient(cfg.RiotApiKey, cfg.RiotRatePerSec, cfg.RiotRatePer2Min),
		ddragon: external.NewDDragonClient(),
	}, nil
}
