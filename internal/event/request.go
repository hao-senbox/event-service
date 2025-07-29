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
