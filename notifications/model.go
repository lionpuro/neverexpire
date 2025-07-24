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

type AppNotification struct {
	ID           int              `db:"id"`
	UserID       string           `db:"user_id"`
	HostID       int              `db:"host_id"`
	Type         NotificationType `db:"notification_type"`
	Body         string           `db:"body"`
	Due          time.Time        `db:"due"`
	DeliveredAt  *time.Time       `db:"delivered_at"`
	ReadAt       *time.Time       `db:"read_at"`
	Attempts     int              `db:"attempts"`
	DeletedAfter time.Time        `db:"deleted_after"`
	CreatedAt    time.Time        `db:"created_at"`
}

type NotificationUpdate struct {
	ID          int        `db:"id"`
	Body        *string    `db:"body"`
	DeliveredAt *time.Time `db:"delivered_at"`
	ReadAt      *time.Time `db:"read_at"`
	Attempts    *int       `db:"attempts"`
}
