package storage

type UserRepository interface {
	GetUserByEmail(email string) (*User, error)
	CreateUser(user *User) (*User, error)
}
