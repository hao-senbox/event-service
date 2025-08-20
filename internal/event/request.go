package event

type CreateEventRequest struct {
	UserID    string `json:"user_id"`
	EventName string `json:"event_name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`

	IsShow bool `json:"is_show"`
	IsSend bool `json:"is_send"`

	SoundKey         string `bson:"sound_key" json:"sound_key"`
	SoundRepeatTimes int64  `bson:"sound_repeat_times" json:"sound_repeat_times"`
	Icon             string `json:"icon"`
	Note             string `json:"note"`
	Url              string `json:"url"`

	Reminders []ReminderRule   `json:"reminder_settings"`
	Schedule  ScheduleSettings `json:"scheduled_settings"`
}

type TriggerEventRequest struct {
	EventID string `json:"event_id"`
}

type UpdateEventRequest struct {
	EventName *string `json:"event_name,omitempty"`
	StartDate *string `json:"start_date,omitempty"`
	EndDate   *string `json:"end_date,omitempty"`

	IsShow *bool `json:"is_show,omitempty"`
	IsSend *bool `json:"is_send,omitempty"`

	SoundKey         *string `bson:"sound_key" json:"sound_key"`
	SoundRepeatTimes *int64  `bson:"sound_repeat_times" json:"sound_repeat_times"`
	Icon             *string `json:"icon,omitempty"`
	Note             *string `json:"note,omitempty"`
	Url              *string `json:"url,omitempty"`

	Reminders *[]ReminderRule   `json:"reminder_settings,omitempty"`
	Schedule  *ScheduleSettings `json:"scheduled_settings,omitempty"`
}
