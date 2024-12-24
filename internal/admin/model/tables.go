package model

type AdminSchema struct {
	Username   string `gorm:"column:username;type:varchar(255)" json:"username"`
	Password   string `gorm:"column:password;type:varchar(255)" json:"password"`
	Email      string `gorm:"column:email;type:varchar(255)" json:"email"`
	TelegramID string `gorm:"column:telegram_id;type:varchar(255)" json:"telegram_id"`
}
