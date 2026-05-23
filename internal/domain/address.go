package domain

type Address struct {
	ID          int64  `json:"id" db:"id"`
	UserID      int64  `json:"user_id" db:"user_id"`
	StreetHouse string `json:"street_house" db:"street_house"`
	Apartment   string `json:"apartment" db:"apartment"`
	Entrance    string `json:"entrance" db:"entrance"`
	Floor       string `json:"floor" db:"floor"`
	DoorCode    string `json:"door_code" db:"door_code"`
}
