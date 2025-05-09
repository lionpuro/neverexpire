package model

type User struct {
	ID    string `db:"id"`
	Email string `db:"email"`
}

type SessionUser struct {
	ID    string
	Email string
}
