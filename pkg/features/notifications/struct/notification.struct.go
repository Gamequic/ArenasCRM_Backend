package notificatonstruct

type NotificationCreate struct {
	UserId  int    `validate:"required"`
	Message string `validate:"required"`
}
