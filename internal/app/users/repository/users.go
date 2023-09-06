package repository

import (
	"context"
	"github.com/exclide/movie-service/internal/app/model"
	"github.com/exclide/movie-service/internal/app/store"
)

type UserRepository struct {
	Store *store.Store
}

func NewUserRepository(s *store.Store) UserRepository {
	return UserRepository{s}
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) (*model.User, error) {
	stmt, err := r.Store.Db.Prepare("INSERT INTO users (login, password) VALUES ($1, $2)")
	if err != nil {
		return nil, err
	}

	_, err = stmt.Exec(u.Login, u.Password)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	var mv model.User

	stmt, err := r.Store.Db.Prepare("select * from users where login = $1")
	if err != nil {
		return nil, err
	}

	err = stmt.QueryRow(login).Scan(&mv.Login, &mv.Password)
	if err != nil {
		return nil, err
	}

	return &mv, nil
}

func (r *UserRepository) DeleteByLogin(ctx context.Context, login string) error {
	stmt, err := r.Store.Db.Prepare("delete from users where login = $1")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(login)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) GetAll(ctx context.Context) ([]model.User, error) {
	var mvs []model.User

	stmt, err := r.Store.Db.Prepare("select * from users")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	defer rows.Close()

	for rows.Next() {
		var mv model.User
		if err := rows.Scan(&mv.Login, &mv.Password); err != nil {
			return nil, err
		}
		mvs = append(mvs, mv)
	}

	return mvs, nil
}