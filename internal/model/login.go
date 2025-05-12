package model

type UserLogin struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (ul *UserLogin) Validate() error {
	return validate.Struct(ul)
}
