package dto

type UserWithJwt struct {
	Name   string
	Email  string
	Avatar string
	Token  string
}

type RegisterUserDto struct {
	Name     string
	Email    string
	Password string
}
