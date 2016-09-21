package models

const (
	WebHookNormal = itoa
	WebHookSlack
)

type WebHook struct {
	ID          int        `gorm:"primary_key" json:"id"`
	Type        int        `json:"type"`
	DeliveryURL string     "json:deliveryURL"
	Headers     string     `json:"headers"`
	Content     string     `json:"content"`
	Projects    []*Project `gorm:"many2many:project_webhooks" json:"-"`
}
