package entity

type User struct {
	id             string
	Login          string
	HashedPassword string
}

func NewUser(login, hashedPassword string) *User {
	return &User{Login: login, HashedPassword: hashedPassword}
}
