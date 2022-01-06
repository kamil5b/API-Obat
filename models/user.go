package models

type User struct {
	Id       uint   `json:"id" gorm:"unique"`
	Name     string `json:"name"`
	NIK      string `json:"nik" gorm:"unique"`
	Password []byte `json:"-"`
	Role     string `json:"role"`
}
