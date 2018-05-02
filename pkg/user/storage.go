package user

import (
	"github.com/jackc/pgx"
)

// Storage provider that can handle read/write operation to database/file/bytes
type Storage interface {
	GetUserByEmail(*User) error
	CreateUser(*User) error
}

// PGStorage provider that can handle read/write from database
type PGStorage struct {
	con *pgx.ConnPool
}

// NewPostgres will open db connection or return error
func NewPostgres(con *pgx.ConnPool) (pg *PGStorage) {

	pg = &PGStorage{
		con: con,
	}
	return pg
}

// CreateUser pull user from postgresql database
func (pg *PGStorage) CreateUser(u *User) error {

	err := pg.con.QueryRow("INSERT INTO users(email, password) VALUES($1, $2) RETURNING id",
		u.Email,
		u.Password,
	).Scan(&u.ID)

	return err
}

// GetUserByEmail pull user from postgresql database
func (pg *PGStorage) GetUserByEmail(u *User) (err error) {
	err = pg.con.QueryRow("SELECT id, password FROM users WHERE email=$1 and password=$2",
		u.Email, u.Password,
	).Scan(&u.ID, &u.Password)

	return err
}
