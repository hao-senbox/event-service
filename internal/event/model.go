package event

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event struct {
	ID              primitive.ObjectID `bson:"_id" json:"id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	EventName       string             `bson:"event_name" json:"event_name"`
	StartDate       time.Time          `bson:"start_date" json:"start_date"`
	EndDate         time.Time          `bson:"end_date" json:"end_date"`
	IsShow          bool               `bson:"is_show" json:"is_show"`
	IsSend          bool               `bson:"is_send" json:"is_send"`
	SoundKey        string             `bson:"sound_key" json:"sound_key"`
	SoundRepeatTimes int64              `bson:"sound_repeat_times" json:"sound_repeat_times"`
	Icon            string             `bson:"icon" json:"icon"`
	Note            string             `bson:"note" json:"note"`
	Url             string             `bson:"url" json:"url"`
	Reminders       []ReminderRule          `bson:"reminder_settings" json:"reminder_settings"`
	Schedule        ScheduleSettings   `bson:"scheduled_settings" json:"scheduled_settings"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type ReminderRule struct {
	RemiderCount   int64   `bson:"remider_count" json:"remider_count"`
	ReminderBefore string  `bson:"reminder_before" json:"reminder_before"`
	Enable         bool    `bson:"enable" json:"enable"`
	Message        *string `bson:"message,omitempty" json:"message,omitempty"`
}

type DayOption struct {
	Key   string `bson:"key" json:"key"`
	Value string `bson:"value" json:"value"`
}

type ScheduleSettings struct {
	Day        []DayOption `bson:"day_selections" json:"day_selections"`
	Expiration int         `bson:"expiration" json:"expiration"`
}
