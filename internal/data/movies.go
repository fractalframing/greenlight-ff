package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/fractalframing/greenlight-ff/internal/validator"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

type MovieModel struct {
	DB *sql.DB
}

func (m MovieModel) Insert(movie *Movie) error {
	insertQuery := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES (?, ?, ?, ?);
	`
	selectQuery := `
		SELECT LAST_INSERT_ID() as id, created_at, version
		FROM movies
		WHERE id = LAST_INSERT_ID();
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []any{movie.Title, movie.Year, movie.Runtime, strings.Join(movie.Genres, ",")}
	_, err := m.DB.ExecContext(ctx, insertQuery, args...)
	if err != nil {
		return err
	}
	row := m.DB.QueryRowContext(ctx, selectQuery)
	if err := row.Scan(&movie.ID, &movie.CreatedAt, &movie.Version); err != nil {
		return err
	}
	return nil
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE id = ?;
	`
	tmp := struct {
		ID        int64
		CreatedAt time.Time
		Title     string
		Year      int32
		Runtime   int32
		Genres    string
		Version   int32
	}{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&tmp.ID,
		&tmp.CreatedAt,
		&tmp.Title,
		&tmp.Year,
		&tmp.Runtime,
		&tmp.Genres,
		&tmp.Version,
	)
	movie := Movie{
		ID:        tmp.ID,
		CreatedAt: tmp.CreatedAt,
		Title:     tmp.Title,
		Year:      tmp.Year,
		Runtime:   Runtime(tmp.Runtime),
		Genres:    strings.Split(tmp.Genres, ","),
		Version:   tmp.Version,
	}
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {
	updateQuery := `
		UPDATE movies
		SET title = ?, year = ?, runtime = ?, genres = ?, version = version + 1
		WHERE id = ? AND version = ?;
	`
	selectQuery := `
		SELECT version
		FROM movies
		WHERE id = ?;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []any{movie.Title, movie.Year, movie.Runtime, strings.Join(movie.Genres, ","), movie.ID, movie.Version}
	_, err := m.DB.ExecContext(ctx, updateQuery, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	row := m.DB.QueryRowContext(ctx, selectQuery, movie.ID)
	if err := row.Scan(&movie.Version); err != nil {
		return err
	}
	return nil
}

func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
		DELETE FROM movies
		WHERE id=?;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
