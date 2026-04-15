package storage

import (
	"context"
	"github.com/uptrace/bun"
)

type UserRepositoryImpl struct {
	Db *bun.DB
}

func NewUserRepositoryImpl(db *bun.DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{Db: db}
}

func (u UserRepositoryImpl) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := u.Db.NewSelect().Model(user).Where("email = ?", email).Scan(context.Background())
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u UserRepositoryImpl) CreateUser(user *User) (*User, error) {
	//TODO implement me
	panic("implement me")
}
