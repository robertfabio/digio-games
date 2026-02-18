package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Save struct {
	ID        int       `json:"id"`
	RomName   string    `json:"rom_name"`
	SaveType  string    `json:"save_type"`
	Slot      int       `json:"slot"`
	Data      []byte    `json:"-"`
	Size      int       `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS saves (
			id         SERIAL PRIMARY KEY,
			rom_name   TEXT NOT NULL,
			save_type  TEXT NOT NULL DEFAULT 'sram',
			slot       INT NOT NULL DEFAULT 0,
			data       BYTEA NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(rom_name, save_type, slot)
		)
	`)
	return err
}

func (s *Store) ListSaves(ctx context.Context, romName string) ([]Save, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, rom_name, save_type, slot, length(data), created_at, updated_at
		FROM saves WHERE rom_name = $1 ORDER BY updated_at DESC
	`, romName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var saves []Save
	for rows.Next() {
		var sv Save
		if err := rows.Scan(&sv.ID, &sv.RomName, &sv.SaveType, &sv.Slot, &sv.Size, &sv.CreatedAt, &sv.UpdatedAt); err != nil {
			return nil, err
		}
		saves = append(saves, sv)
	}
	return saves, rows.Err()
}

func (s *Store) GetSaveData(ctx context.Context, romName, saveType string, slot int) ([]byte, error) {
	var data []byte
	err := s.pool.QueryRow(ctx, `
		SELECT data FROM saves
		WHERE rom_name = $1 AND save_type = $2 AND slot = $3
	`, romName, saveType, slot).Scan(&data)
	return data, err
}

func (s *Store) UpsertSave(ctx context.Context, romName, saveType string, slot int, data []byte) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO saves (rom_name, save_type, slot, data)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (rom_name, save_type, slot)
		DO UPDATE SET data = EXCLUDED.data, updated_at = NOW()
	`, romName, saveType, slot, data)
	return err
}

func (s *Store) DeleteSave(ctx context.Context, id int) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM saves WHERE id = $1`, id)
	return err
}
