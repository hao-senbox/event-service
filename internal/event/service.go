package event

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"event-service/internal/user"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventService interface {
	CreateEvent(ctx context.Context, req *CreateEventRequest) error
	GetAllEvents(ctx context.Context, userID string) ([]*Event, error)
	GetEventByID(ctx context.Context, eventID string) (*Event, error)
	UpdateEvent(ctx context.Context, req *UpdateEventRequest, id string) error
	DeleteEvent(ctx context.Context, id string) error
	ToggleSendEventNotifications(ctx context.Context, id string) (string, error)
	ToggleShowEventNotifications(ctx context.Context, id string) (string, error)
	CronEventNotifications(ctx context.Context) error
	SendEventNotifications(ctx context.Context, req *TriggerEventRequest) error
}

type eventService struct {
	eventRepository EventRepository
	fireBase        *firebase.App
	userService     user.UserService
	location        *time.Location
}

func NewEventService(repo EventRepository, fb *firebase.App, us user.UserService) EventService {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		loc = time.UTC
	}
	return &eventService{
		eventRepository: repo,
		fireBase:        fb,
		userService:     us,
		location:        loc,
	}
}

func (s *eventService) CreateEvent(ctx context.Context, req *CreateEventRequest) error {

	if req.UserID == "" || req.EventName == "" {
		return errors.New("user_id and event_name are required")
	}

	if req.StartDate == "" || req.EndDate == "" {
		return errors.New("start_date and end_date are required")
	}

	start, err := time.ParseInLocation("2006-01-02 15:04:05", req.StartDate, s.location)
	if err != nil {
		return fmt.Errorf("invalid start_date: %w", err)
	}

	end, err := time.ParseInLocation("2006-01-02 15:04:05", req.EndDate, s.location)
	if err != nil {
		return fmt.Errorf("invalid end_date: %w", err)
	}

	if end.Before(start) {
		return errors.New("end_date must be after start_date")
	}

	if req.Schedule.Expiration < 0 {
		req.Schedule.Expiration = 0
	}

	ev := &Event{
		ID:               primitive.NewObjectID(),
		UserID:           req.UserID,
		EventName:        req.EventName,
		StartDate:        start.In(s.location),
		EndDate:          end.In(s.location),
		IsShow:           true,
		IsSend:           true,
		Reminders:        req.Reminders,
		Schedule:         req.Schedule,
		Note:             req.Note,
		SoundKey:         req.SoundKey,
		SoundRepeatTimes: req.SoundRepeatTimes,
		Icon:             req.Icon,
		Url:              req.Url,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return s.eventRepository.Create(ctx, ev)
}

func (s *eventService) UpdateEvent(ctx context.Context, req *UpdateEventRequest, id string) error {

	if id == "" {
		return errors.New("event_id is required")
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	ev, err := s.eventRepository.FindEventByID(ctx, objID)
	if err != nil || ev == nil {
		return errors.New("event not found")
	}

	if req.EventName != nil {
		ev.EventName = *req.EventName
	}

	if req.StartDate != nil {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", *req.StartDate, s.location)
		if err != nil {
			return fmt.Errorf("invalid start_date: %w", err)
		}
		ev.StartDate = t.In(s.location)
	}

	if req.EndDate != nil {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", *req.EndDate, s.location)
		if err != nil {
			return fmt.Errorf("invalid end_date: %w", err)
		}
		ev.EndDate = t.In(s.location)
	}

	if req.IsShow != nil {
		ev.IsShow = *req.IsShow
	}

	if req.IsSend != nil {
		ev.IsSend = *req.IsSend
	}

	if req.Reminders != nil {
		ev.Reminders = *req.Reminders
	}

	if req.Schedule != nil {
		ev.Schedule = *req.Schedule
		if ev.Schedule.Expiration < 0 {
			ev.Schedule.Expiration = 0
		}
	}

	if req.Note != nil {
		ev.Note = *req.Note
	}

	if req.SoundKey != nil {
		ev.SoundKey = *req.SoundKey
		ev.SoundRepeatTimes = *req.SoundRepeatTimes
	}

	if req.Icon != nil {
		ev.Icon = *req.Icon
	}

	if req.Url != nil {
		ev.Url = *req.Url
	}

	if ev.StartDate.After(ev.EndDate) {
		return errors.New("end_date must be after start_date")
	}

	ev.UpdatedAt = time.Now()

	return s.eventRepository.UpdateEvent(ctx, ev, objID)

}

func (s *eventService) CronEventNotifications(ctx context.Context) error {

	now := time.Now().In(s.location).Truncate(time.Minute)

	log.Printf("🕐 Cron check at: %s", now.Format("2006-01-02 15:04:05"))

	events, err := s.eventRepository.FindEventActive(ctx)
	if err != nil {
		log.Printf("❌ Error FindEventActive: %v", err)
		return err
	}

	log.Printf("📋 Found %d active events", len(events))

	for _, ev := range events {
		start := ev.StartDate.In(s.location)
		end := ev.EndDate.In(s.location)
		log.Printf("➡️ Checking event %s (Start=%s, End=%s, Expiration=%d, StartHHmm=%02d:%02d)",
			ev.EventName,
			start.Format("2006-01-02 15:04:05"),
			end.Format("2006-01-02 15:04:05"),
			ev.Schedule.Expiration,
			start.Hour(), start.Minute(),
		)

		if s.shouldSendNotification(ev, now) {
			log.Printf("✅ Triggered event: %s", ev.EventName)
			s.sendNotification(ctx, ev)
		} else {
			log.Printf("⏭️ Skipped event: %s", ev.EventName)
		}
	}

	return nil
}

func (s *eventService) shouldSendNotification(ev *Event, now time.Time) bool {

	log.Printf("🔍 shouldSendNotification: now=%s", now.Format("2006-01-02 15:04:05"))

	if !ev.IsSend || !ev.IsShow {
		log.Printf("⛔ Event disabled (IsSend=%t, IsShow=%t)", ev.IsSend, ev.IsShow)
		return false
	}

	start := ev.StartDate.In(s.location)
	end := ev.EndDate.In(s.location)

	if now.After(end) {
		log.Printf("⛔ now > EndDate: now=%s, End=%s",
			now.Format("2006-01-02 15:04:05"),
			end.Format("2006-01-02 15:04:05"))
		return false
	}

	exp := ev.Schedule.Expiration
	if exp < 0 {
		exp = 0
	}

	repeats := exp
	if repeats == 0 {
		repeats = 1
	}

	interval := time.Minute

	startHH, startMM := start.Hour(), start.Minute()

	for ridx, rule := range ev.Reminders.Rules {

		if !rule.Enable {
			log.Printf("⏭️ Rule %d inactive (offset=%d %s)", ridx, rule.RemiderCount, rule.ReminderBefore)
			continue
		}

		for k := 0; k < repeats; k++ {

			base := now.Add(-time.Duration(k) * interval)
			occCandidate := s.addOffset(base, rule)

			occ := time.Date(occCandidate.Year(), occCandidate.Month(), occCandidate.Day(),
				startHH, startMM, 0, 0, s.location)

			log.Printf("🧮 Rule %d, k=%d -> occ=%s (weekday=%s); window=[%s..%s]",
				ridx, k,
				occ.Format("2006-01-02 15:04:05"),
				occ.Weekday().String(),
				start.Format("2006-01-02 15:04:05"),
				end.Format("2006-01-02 15:04:05"),
			)

			if occ.Before(start) || occ.After(end) {
				log.Printf("   ⛔ occ out of range")
				continue
			}

			if len(ev.Schedule.Day) > 0 && !s.weekdayAllowed(occ.Weekday(), ev.Schedule.Day) {
				log.Printf("   ⛔ occ weekday %s not allowed by Day selection", occ.Weekday())
				continue
			}

			target := s.subtractOffset(occ, rule).Truncate(time.Minute)
			candidate := target.Add(time.Duration(k) * interval) // target + k*1'
			log.Printf("🎛️ Check rule=%d k=%d: target=%s | candidate=%s | now=%s",
				ridx, k,
				target.Format("2006-01-02 15:04:05"),
				candidate.Format("2006-01-02 15:04:05"),
				now.Format("2006-01-02 15:04:05"),
			)

			if now.Equal(candidate) {
				log.Printf("🎯 EXACT MATCH → send (rule=%d, k=%d)", ridx, k)
				return true
			}
		}
	}
	return false
}

func (s *eventService) subtractOffset(base time.Time, r ReminderRule) time.Time {
	switch r.ReminderBefore {
	case "minutes":
		return base.Add(-time.Duration(r.RemiderCount) * time.Minute)
	case "hours":
		return base.Add(-time.Duration(r.RemiderCount) * time.Hour)
	case "days":
		return base.AddDate(0, 0, -int(r.RemiderCount))
	case "weeks":
		return base.AddDate(0, 0, -int(r.RemiderCount*7))
	case "months":
		return base.AddDate(0, -int(r.RemiderCount), 0)
	default:
		return base.Add(-time.Duration(r.RemiderCount) * time.Minute)
	}
}

func (s *eventService) weekdayAllowed(wd time.Weekday, days []DayOption) bool {
	key := strings.ToLower(wd.String())
	for _, d := range days {
		if strings.ToLower(d.Key) == key {
			return true
		}
	}
	return false
}

func (s *eventService) addOffset(base time.Time, r ReminderRule) time.Time {
	switch r.ReminderBefore {
	case "minutes":
		return base.Add(time.Duration(r.RemiderCount) * time.Minute)
	case "hours":
		return base.Add(time.Duration(r.RemiderCount) * time.Hour)
	case "days":
		return base.AddDate(0, 0, int(r.RemiderCount))
	case "weeks":
		return base.AddDate(0, 0, int(r.RemiderCount*7))
	case "months":
		return base.AddDate(0, int(r.RemiderCount), 0)
	default:
		return base.Add(time.Duration(r.RemiderCount) * time.Minute)
	}
}

func (s *eventService) sendNotification(ctx context.Context, event *Event) {

	tokens, err := s.userService.GetTokenUser(ctx, event.UserID)
	if err != nil || tokens == nil {
		log.Printf("❌ GetTokenUser error for user %s: %v", event.UserID, err)
		return
	}

	if len(*tokens) == 0 {
		log.Printf("📵 No tokens found for user %s", event.UserID)
		return
	}

	message := s.getNotificationMessage(event)

	successCount := 0
	for _, token := range *tokens {
		if token == "" {
			continue
		}

		client, err := s.fireBase.Messaging(ctx)
		if err != nil {
			log.Printf("❌ Firebase client error: %v", err)
			continue
		}

		msg := &messaging.Message{
			Notification: &messaging.Notification{
				Title: "🔔 " + event.EventName,
				Body:  message,
			},
			Token: token,
		}

		response, err := client.Send(ctx, msg)
		if err != nil {
			log.Printf("❌ Failed to send to token %s: %v", token, err)
		} else {
			log.Printf("✅ Sent notification to token %s (response: %s)", token, response)
			successCount++
		}
	}

	log.Printf("📊 Event %s: sent %d/%d notifications successfully", event.EventName, successCount, len(*tokens))
}

func (s *eventService) getNotificationMessage(event *Event) string {
	return fmt.Sprintf("Nhắc nhở: %s sắp bắt đầu!", event.EventName)
}

func (s *eventService) GetAllEvents(ctx context.Context, userID string) ([]*Event, error) {

	if userID == "" {
		return nil, errors.New("user_id is required")
	}

	return s.eventRepository.FindAllEvents(ctx, userID)
}

func (s *eventService) GetEventByID(ctx context.Context, eventID string) (*Event, error) {

	if eventID == "" {
		return nil, errors.New("event_id is required")
	}

	objID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		return nil, err
	}

	return s.eventRepository.FindEventByID(ctx, objID)

}

func (s *eventService) DeleteEvent(ctx context.Context, id string) error {

	if id == "" {
		return errors.New("event_id is required")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	return s.eventRepository.DeleteEvent(ctx, objectID)

}

func (s *eventService) ToggleSendEventNotifications(ctx context.Context, id string) (string, error) {

	var check string
	if id == "" {
		return "", errors.New("event_id is required")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return "", err
	}

	event, err := s.eventRepository.FindEventByID(ctx, objectID)
	if err != nil {
		return "", err
	}

	if event == nil {
		return "", errors.New("event not found")
	}

	if event.IsSend {
		event.IsSend = false
		check = "off"
	} else {
		event.IsSend = true
		check = "on"
	}

	err = s.eventRepository.UpdateEvent(ctx, event, objectID)
	if err != nil {
		return "", err
	}

	return check, nil
}

func (s *eventService) ToggleShowEventNotifications(ctx context.Context, id string) (string, error) {

	var check string

	if id == "" {
		return "", errors.New("event_id is required")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return "", err
	}

	event, err := s.eventRepository.FindEventByID(ctx, objectID)
	if err != nil {
		return "", err
	}

	if event == nil {
		return "", errors.New("event not found")
	}

	if event.IsShow {
		event.IsShow = false
		check = "off"
	} else {
		event.IsShow = true
		check = "on"
	}

	err = s.eventRepository.UpdateEvent(ctx, event, objectID)
	if err != nil {
		return "", err
	}

	return check, nil
}

func (s *eventService) SendEventNotifications(ctx context.Context, req *TriggerEventRequest) error {

	if req.EventID == "" {
		return errors.New("event_id is required")
	}

	objectID, err := primitive.ObjectIDFromHex(req.EventID)
	if err != nil {
		return err
	}

	event, err := s.eventRepository.FindEventByID(ctx, objectID)
	if err != nil {
		return err
	}

	if event == nil {
		return errors.New("event not found")
	}

	s.sendNotification(ctx, event)

	return nil
}
