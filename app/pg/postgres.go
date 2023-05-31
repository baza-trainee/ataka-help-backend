package pg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/baza-trainee/ataka-help-backend/app/config"
	"github.com/baza-trainee/ataka-help-backend/app/structs"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const (
	ErrCodeUniqueViolation     = "unique_violation"
	ErrCodeNoData              = "no_data"
	ErrCodeForeignKeyViolation = "foreign_key_violation"
	ErrCodeUndefinedColumn     = "undefined_column"
)

type Repo struct {
	db sqlx.DB
}

// Returns an object of the Ropository.
func NewRepository(cfg config.Config) (Repo, error) {
	database, err := sqlx.Connect("postgres", fmt.Sprintf("host=%v port=%v user=%v dbname=%v password=%v sslmode=%v",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Database, cfg.DB.Password, cfg.DB.SSLmode))
	if err != nil {
		return Repo{}, fmt.Errorf("cannot connect to db: %w", err)
	}

	return Repo{db: *database}, nil
}

func (r Repo) Close() error {
	return r.db.Close()
}

func (r Repo) SelectAllCards(params structs.CardQueryParameters, ctx context.Context) ([]structs.Card, error) {
	query := `
		SELECT id, title, thumb, alt, description, created, modified
		FROM public.cards c
		ORDER BY c.created DESC
		Limit $1
		OFFSET $2;
	`
	cards := []structs.Card{}

	rows, err := r.db.QueryContext(ctx, query, params.Limit, params.Offset)
	if err != nil {
		return nil, fmt.Errorf("an error occurs while QueryContext: %w", err)
	}

	for rows.Next() {
		card := structs.Card{}

		if err := rows.Scan(
			&card.ID, &card.Title, &card.Thumb, &card.Alt,
			&card.Description, &card.Created, &card.Modified); err != nil {
			return nil, fmt.Errorf("an error occurs while rows.Scan: %w", err)
		}

		cards = append(cards, card)

	}

	return cards, nil
}

func (r Repo) InsertCard(card structs.Card, ctx context.Context) error {
	const expectedEffectedRow = 1

	query := `INSERT INTO public.cards
	(title, thumb, alt, description)
	VALUES($1, $2, $3, $4::jsonb);`

	jsonDescription, err := json.Marshal(card.Description)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, card.Title, card.Thumb, card.Alt, jsonDescription)
	if err != nil {
		pqError := new(pq.Error)
		if errors.As(err, &pqError) && pqError.Code.Name() == ErrCodeUniqueViolation {
			return structs.ErrUniqueRestriction
		}

		return fmt.Errorf("error in NamedEx: %w", err)
	}

	effectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("the error is in RowsAffected: %w", err)
	}

	if effectedRows != expectedEffectedRow {
		return structs.ErrDatabaseInserting
	}

	return nil
}

func (r Repo) CountRowsTable(table string, ctx context.Context) (int, error) {
	query := `SELECT count(*) as result FROM public.` + table
	var total int
	r.db.GetContext(ctx, &total, query)

	fmt.Println(total)
	return total, nil
}

func (r Repo) SelectAllPartners() (string, error) {
	return "some partners from db", nil
}