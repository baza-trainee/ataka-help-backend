package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/baza-trainee/ataka-help-backend/app/config"
	"github.com/baza-trainee/ataka-help-backend/app/structs"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const (
	id                         = "id"
	expectedAffectedRow        = 1
	ErrCodeUniqueViolation     = "unique_violation"
	ErrCodeNoData              = "no_data"
	ErrCodeForeignKeyViolation = "foreign_key_violation"
	ErrCodeUndefinedColumn     = "undefined_column"
)

type Repo struct {
	db *sqlx.DB
}

// Returns an object of the Ropository.
func NewRepository(cfg config.Config) (Repo, error) {
	database, err := sqlx.Connect("postgres", fmt.Sprintf("host=%v port=%v user=%v dbname=%v password=%v sslmode=%v",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Database, cfg.DB.Password, cfg.DB.SSLmode))
	if err != nil {
		return Repo{}, fmt.Errorf("cannot connect to db: %w", err)
	}

	return Repo{db: database}, nil
}

func (r Repo) Close() error {
	return fmt.Errorf("error hapens while db.close: %w", r.db.Close())
}

func (r Repo) UpdateContact(ctx context.Context, contact structs.Contact) error {
	const expectedEffectedRow = 1

	query := `UPDATE public.contact
		SET phone1=$1, phone2=$2, email=$3;
		`
	fmt.Println(contact)
	result, err := r.db.ExecContext(ctx, query, contact.Phone1, contact.Phone2, contact.Email)
	if err != nil { //nolint: wsl
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
		return structs.ErrNoRowAffected
	}

	return nil
}

func (r Repo) SelectContact(ctx context.Context) (structs.Contact, error) {
	query := `SELECT * FROM public.contact LIMIT 1`

	contact := structs.Contact{}

	if err := r.db.GetContext(ctx, &contact, query); err != nil {
		if err == sql.ErrNoRows {
			return structs.Contact{}, structs.ErrNotFound
		}

		return structs.Contact{}, fmt.Errorf("error while GetContext(): %w", err)
	}

	return contact, nil
}

func (r Repo) CountRowsTable(ctx context.Context, table string) (int, error) {
	query := `SELECT count(*) as result FROM public.` + table

	var total int

	if err := r.db.GetContext(ctx, &total, query); err != nil {
		return 0, fmt.Errorf("error in GetContext: %w", err)
	}

	return total, nil
}

func (r Repo) SelectAllPartners(ctx context.Context, params structs.PartnerQueryParameters) ([]structs.Partner, error) {
	partners := []structs.Partner{}

	var limit, page *int = &params.Limit, &params.Page

	if params.Limit == 0 {
		limit = nil
		page = nil
	}

	query := `SELECT id, alt, thumb, created, modified
				FROM public.partners as p
				ORDER BY p.created ASC
				LIMIT $1
				OFFSET $2;`

	rows, err := r.db.QueryContext(ctx, query, limit, page)
	if err != nil {
		return nil, fmt.Errorf("an error occurs while QueryContext: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		partner := structs.Partner{}

		if err := rows.Scan(&partner.ID, &partner.Alt, &partner.Thumb, &partner.Created, &partner.Modified); err != nil {
			return nil, fmt.Errorf("an error occurs while rows.Scan(): %w", err)
		}

		partners = append(partners, partner)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("an error occurs while rows.Err(): %w", err)
	}

	return partners, nil
}

func (r Repo) InsertPartner(ctx context.Context, partner structs.Partner, chWell chan struct{}) error {
	query := `INSERT INTO public.partners (alt, thumb)
			  VALUES($1, $2);`

	result, err := r.db.ExecContext(ctx, query, partner.Alt, partner.Thumb)
	if err != nil {
		pqError := new(pq.Error)
		if errors.As(err, &pqError) && pqError.Code.Name() == ErrCodeForeignKeyViolation {
			return structs.ErrForeignViolation
		}

		if errors.As(err, &pqError) && pqError.Code.Name() == ErrCodeUniqueViolation {
			return structs.ErrUniqueRestriction
		}

		return fmt.Errorf("error in ExecContext: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("the error is in RowsAffected: %w", err)
	}

	if affectedRows != expectedAffectedRow {
		return structs.ErrNoRowAffected
	}

	chWell <- struct{}{}

	close(chWell)

	return nil
}

func (r Repo) SelectSlider(ctx context.Context) ([]structs.Slide, error) {
	records := []structs.Slide{}

	query := `SELECT id, title, thumb, alt, created, modified 
			  FROM public.slider
			  ORDER BY created DESC;`

	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error in QueryxContext: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		record := structs.Slide{}

		err := rows.StructScan(&record)
		if err != nil {
			return nil, fmt.Errorf("error in QueryxContext.Next(): %w", err)
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error in QueryxContext.Err(): %w", err)
	}

	return records, nil
}

func (r Repo) InsertSlider(ctx context.Context, slider structs.Slide, chWell chan struct{}) error {
	query := `INSERT INTO public.slider (title, thumb, alt)
			  VALUES($1, $2, $3);`

	result, err := r.db.ExecContext(ctx, query, slider.Title, slider.Thumb, slider.Alt)
	if err != nil {
		pqError := new(pq.Error)
		if errors.As(err, &pqError) && pqError.Code.Name() == ErrCodeForeignKeyViolation {
			return structs.ErrForeignViolation
		}

		if errors.As(err, &pqError) && pqError.Code.Name() == ErrCodeUniqueViolation {
			return structs.ErrUniqueRestriction
		}

		return fmt.Errorf("error in ExecContext: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("the error is in RowsAffected: %w", err)
	}

	if affectedRows != expectedAffectedRow {
		return structs.ErrNoRowAffected
	}

	chWell <- struct{}{}

	close(chWell)

	return nil
}

func (r Repo) DelSlideByID(ctx context.Context, ID string) (string, error) {
	getQuery := `SELECT thumb FROM public.slider WHERE id = $1`

	objectPath := struct {
		Thumb string `db:"thumb"`
	}{}

	if err := r.db.GetContext(ctx, &objectPath, getQuery, ID); err != nil {
		return "", fmt.Errorf("error in GetContext(): %w", err)
	}

	deleteQuery := `DELETE FROM public.slider WHERE id=$1`

	sqlResult, err := r.db.ExecContext(ctx, deleteQuery, ID)
	if err != nil {
		return "", fmt.Errorf("error in ExecContext(): %w", err)
	}

	affectedRows, err := sqlResult.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("the error is in RowsAffected: %w", err)
	}

	if affectedRows != expectedAffectedRow {
		return "", structs.ErrNoRowAffected
	}

	return objectPath.Thumb, nil
}

func (r Repo) DelPartnerByID(ctx context.Context, ID string) (string, error) {
	getQuery := `SELECT thumb FROM public.partners WHERE id = $1`

	objectPath := struct {
		Thumb string `db:"thumb"`
	}{}

	if err := r.db.GetContext(ctx, &objectPath, getQuery, ID); err != nil {
		return "", fmt.Errorf("error in GetContext(): %w", err)
	}

	deleteQuery := `DELETE FROM public.partners WHERE id=$1`

	sqlResult, err := r.db.ExecContext(ctx, deleteQuery, ID)
	if err != nil {
		return "", fmt.Errorf("error in ExecContext(): %w", err)
	}

	affectedRows, err := sqlResult.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("the error is in RowsAffected: %w", err)
	}

	if affectedRows != expectedAffectedRow {
		return "", structs.ErrNoRowAffected
	}

	return objectPath.Thumb, nil
}
