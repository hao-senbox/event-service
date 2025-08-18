package event

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	UserID    string             `bson:"user_id" json:"user_id"`
	EventName string             `bson:"event_name" json:"event_name"`
	StartDate time.Time          `bson:"start_date" json:"start_date"`
	EndDate   time.Time          `bson:"end_date" json:"end_date"`
	IsShow    bool               `bson:"is_show" json:"is_show"`
	IsSend    bool               `bson:"is_send" json:"is_send"`
	Reminders Reminders          `bson:"reminders" json:"reminders"`
	Schedule  ScheduleSettings   `bson:"schedule" json:"schedule"`
	Media     Media              `bson:"media" json:"media"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type Reminders struct {
	ReminderTime   int64   `bson:"reminder_time" json:"reminder_time"`
	Message        *string `bson:"message" json:"message"`
	ActiveReminder bool    `bson:"active_reminder" json:"active_reminder"`
}

type ScheduleSettings struct {
	Sound      string   `bson:"sound" json:"sound"`
	Repeat     string   `bson:"repeat" json:"repeat"`
	Day        []string `bson:"day" json:"day"`
	Expiration int      `bson:"expiration" json:"expiration"`
}

type Media struct {
	EventIcon string `bson:"event_icon" json:"event_icon"`
	QrCode       string `bson:"qr_code" json:"qr_code"`
	Url          string `bson:"url" json:"url"`
}
