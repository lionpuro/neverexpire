package notifications

import "time"

type NotificationType int

const (
	NotificationTypeExpiration NotificationType = iota
)

func (t NotificationType) String() string {
	switch t {
	case NotificationTypeExpiration:
		return "expiration"
	}
	return ""
}

type Notification struct {
	ID           int              `db:"id"`
	Endpoint     string           `db:"endpoint"`
	UserID       string           `db:"user_id"`
	HostID       int              `db:"host_id"`
	Type         NotificationType `db:"notification_type"`
	Body         string           `db:"body"`
	Due          time.Time        `db:"due"`
	DeliveredAt  *time.Time       `db:"delivered_at"`
	Attempts     int              `db:"attempts"`
	DeletedAfter time.Time        `db:"deleted_after"`
}

type NotificationInput struct {
	UserID       string
	HostID       int
	Type         NotificationType
	Body         string
	Due          time.Time
	DeliveredAt  *time.Time
	Attempts     int
	DeletedAfter time.Time
}

type NotificationUpdate struct {
	UserID       *string
	HostID       *int
	Type         *NotificationType
	Body         *string
	Due          *time.Time
	DeliveredAt  *time.Time
	Attempts     *int
	DeletedAfter *time.Time
}
