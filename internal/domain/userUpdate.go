package domain

type UpdateUserParams struct {
	ID      int64   `json:"-"`
	Name    *string `json:"name"`
	Phone   *string `json:"phone"`
	Email   *string `json:"email"`
	Address *string `json:"address"`
}
