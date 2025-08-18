package event

type CreateEventRequest struct {
	UserID    string           `json:"user_id"`
	EventName string           `json:"event_name"`
	StartDate string           `json:"start_date"`
	EndDate   string           `json:"end_date"`
	Active    bool             `json:"active"`
	Reminders Reminders        `json:"reminders"`
	Schedule  ScheduleSettings `json:"schedule"`
	Media     Media            `json:"media"`
}

type TriggerEventRequest struct {
	EventID   string `json:"event_id"`
}

type UpdateEventRequest struct {
    EventName *string `json:"event_name,omitempty"`
    StartDate *string `json:"start_date,omitempty"` 
    EndDate   *string `json:"end_date,omitempty"`

    IsShow *bool `json:"is_show,omitempty"`
    IsSend *bool `json:"is_send,omitempty"`

    Reminders *struct {
        ReminderTime   *int64 `json:"reminder_time,omitempty"`
        Message        *string `json:"message,omitempty"`
        ActiveReminder *bool   `json:"active_reminder,omitempty"`
    } `json:"reminders,omitempty"`

    Schedule *struct {
        Sound      *string `json:"sound,omitempty"`
        Repeat     *string `json:"repeat,omitempty"`
        Day        *[]string `json:"day,omitempty"` 
        Expiration *int    `json:"expiration,omitempty"` 
    } `json:"schedule,omitempty"`

    Media *struct {
        EventIcon *string `json:"event_icon,omitempty"`
        Url       *string `json:"url,omitempty"`
    } `json:"media,omitempty"`
}
