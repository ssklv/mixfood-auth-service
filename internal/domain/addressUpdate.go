package domain

type UpdateAddressParams struct {
	ID          int64   `json:"-"`
	UserID      int64   `json:"-"`
	StreetHouse *string `json:"street_house"`
	Apartment   *string `json:"apartment"`
	Entrance    *string `json:"entrance"`
	Floor       *string `json:"floor"`
	DoorCode    *string `json:"door_code"`
}
