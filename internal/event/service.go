package event

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"event-service/internal/user"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventService interface {
	CreateEvent(ctx context.Context, req *CreateEventRequest) error
	CronEventNotifications(ctx context.Context) error
	GetAllEvents(ctx context.Context, userID string) ([]*Event, error)
	SendEventNotifications(ctx context.Context, req *TriggerEventRequest) error
}

type eventService struct {
	eventRepository EventRepository
	fireBase        *firebase.App
	userService     user.UserService
	cron            *cron.Cron
	location        *time.Location
}

func NewEventService(repo EventRepository, fb *firebase.App, us user.UserService, cronMaster *cron.Cron) EventService {

	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		log.Printf("⚠️ Failed to load timezone, using UTC: %v", err)
		loc = time.UTC
	}
	
	return &eventService{
		eventRepository: repo,
		fireBase:        fb,
		userService:     us,
		cron:            cronMaster,
		location:        loc,
	}
}

func (s *eventService) CreateEvent(ctx context.Context, req *CreateEventRequest) error {
	if req.UserID == "" {
		return errors.New("user id is required")
	}

	if req.EventName == "" {
		return errors.New("event name is required")
	}

	if req.StartDate == "" || req.EndDate == "" {
		return errors.New("start and end date are required")
	}

	start, err := time.ParseInLocation("2006-01-02 15:04:05", req.StartDate, s.location)
	if err != nil {
		return fmt.Errorf("invalid start date: %v", err)
	}

	end, err := time.ParseInLocation("2006-01-02 15:04:05", req.EndDate, s.location)
	if err != nil {
		return fmt.Errorf("invalid end date: %v", err)
	}

	if end.Before(start) {
		return errors.New("end date must be after start date")
	}

	if req.Schedule.Expiration < 1 || req.Schedule.Expiration > 20 { 
		req.Schedule.Expiration = 1 
	}

	event := &Event{
		ID:        primitive.NewObjectID(),
		UserID:    req.UserID,
		EventName: req.EventName,
		StartDate: start.UTC(), 
		EndDate:   end.UTC(),  
		Active:    true,
		Reminders: Reminders{
			ReminderTime:   req.Reminders.ReminderTime,
			Message:        req.Reminders.Message,
			ActiveReminder: true,
		},
		Schedule: ScheduleSettings{
			Sound:      req.Schedule.Sound,
			Repeat:     req.Schedule.Repeat,
			Day:        req.Schedule.Day,
			Expiration: req.Schedule.Expiration,
		},
		Media: Media{
			EventPicture: req.Media.EventPicture,
			Url:          req.Media.Url,
		},
	}

	return s.eventRepository.Create(ctx, event)
}

func (s *eventService) CronEventNotifications(ctx context.Context) error {
	now := time.Now().In(s.location)
	currentMinute := time.Date(now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), 0, 0, s.location)

	log.Printf("🕐 Checking notifications at: %s (VN time)", currentMinute.Format("2006-01-02 15:04:05"))

	events, err := s.eventRepository.FindEventActive(ctx)
	if err != nil {
		log.Printf("❌ Error getting active events: %v", err)
		return err
	}

	log.Printf("📋 Found %d active events to check", len(events))

	notificationsSent := 0
	for _, event := range events {
		if s.shouldSendNotification(event, currentMinute) {
			log.Printf("📨 Sending notification for event: %s", event.EventName)
			s.sendNotification(ctx, event)
			notificationsSent++
		}
	}

	log.Printf("✅ Sent %d notifications", notificationsSent)
	return nil
}

func (s *eventService) shouldSendNotification(event *Event, currentTime time.Time) bool {
	if !event.Reminders.ActiveReminder {
		log.Printf("⏸️  Event %s: reminder not active", event.EventName)
		return false
	}

	if !event.Active {
		log.Printf("⏸️  Event %s: event not active", event.EventName)
		return false
	}

	minutesBeforeEvent := event.Reminders.ReminderTime
	log.Printf("⏰ Event %s: reminder %d minutes before event", event.EventName, minutesBeforeEvent)

	// Convert stored UTC times to Vietnam timezone for comparison
	startDate := event.StartDate.In(s.location)
	endDate := event.EndDate.In(s.location)
	schedule := event.Schedule

	// Check if current time is within the event's active period
	maxReminderTime := time.Duration(minutesBeforeEvent) * time.Minute
	allowedStartTime := startDate.Add(-maxReminderTime)

	if minutesBeforeEvent == 0 {
		allowedStartTime = startDate
	}

	if currentTime.Before(allowedStartTime) || currentTime.After(endDate) {
		log.Printf("📅 Event %s: outside notification range (%s to %s)",
			event.EventName,
			allowedStartTime.Format("2006-01-02 15:04:05"),
			endDate.Format("2006-01-02 15:04:05"))
		return false
	}

	log.Printf("🔍 Event %s: checking schedule type '%s'", event.EventName, schedule.Repeat)
	switch schedule.Repeat {
	case "hourly":
		return s.checkHourlySchedule(currentTime, minutesBeforeEvent, startDate, schedule, event.EventName)
	case "every_2_hours":
		return s.checkEvery2HoursSchedule(currentTime, minutesBeforeEvent, startDate, schedule, event.EventName)
	case "daily":
		return s.checkDailySchedule(currentTime, minutesBeforeEvent, startDate, schedule, event.EventName)
	case "weekly":
		return s.checkWeeklySchedule(currentTime, minutesBeforeEvent, startDate, schedule, event.EventName)
	case "monthly":
		return s.checkMonthlySchedule(currentTime, minutesBeforeEvent, startDate, schedule, event.EventName)
	case "yearly":
		return s.checkYearlySchedule(currentTime, minutesBeforeEvent, startDate, schedule, event.EventName)
	default:
		log.Printf("❓ Event %s: unknown repeat type '%s'", event.EventName, schedule.Repeat)
		return false
	}
}

func (s *eventService) checkHourlySchedule(currentTime time.Time, minutesBeforeEvent int64, startDate time.Time, schedule ScheduleSettings, eventName string) bool {

	eventStartTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), startDate.Minute(), startDate.Second(), 0, s.location)

	reminderTime := eventStartTime.Add(-time.Duration(minutesBeforeEvent) * time.Minute)
	targetTime := time.Date(reminderTime.Year(), reminderTime.Month(), reminderTime.Day(),
		reminderTime.Hour(), reminderTime.Minute(), 0, 0, s.location)

	log.Printf("🔄 Event %s: hourly check - current hour: %d, event time: %s, reminder time: %s",
		eventName, currentTime.Hour(), eventStartTime.Format("15:04:05"), targetTime.Format("15:04:05"))

	return s.isTimeMatch(currentTime, targetTime, schedule.Expiration, eventName)
}

func (s *eventService) checkEvery2HoursSchedule(currentTime time.Time, minutesBeforeEvent int64, startDate time.Time, schedule ScheduleSettings, eventName string) bool {

	if currentTime.Hour()%2 != 0 {
		return false
	}

	eventStartTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), startDate.Minute(), startDate.Second(), 0, s.location)

	reminderTime := eventStartTime.Add(-time.Duration(minutesBeforeEvent) * time.Minute)
	targetTime := time.Date(reminderTime.Year(), reminderTime.Month(), reminderTime.Day(),
		reminderTime.Hour(), reminderTime.Minute(), 0, 0, s.location)

	log.Printf("🔄 Event %s: every 2 hours check - current hour: %d, event time: %s, reminder time: %s",
		eventName, currentTime.Hour(), eventStartTime.Format("15:04:05"), targetTime.Format("15:04:05"))

	return s.isTimeMatch(currentTime, targetTime, schedule.Expiration, eventName)
}

func (s *eventService) checkDailySchedule(currentTime time.Time, minutesBeforeEvent int64, startDate time.Time, schedule ScheduleSettings, eventName string) bool {
	eventStartTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(),
		startDate.Hour(), startDate.Minute(), startDate.Second(), 0, s.location)
	
	reminderTime := eventStartTime.Add(-time.Duration(minutesBeforeEvent) * time.Minute)
	targetTime := time.Date(reminderTime.Year(), reminderTime.Month(), reminderTime.Day(),
		reminderTime.Hour(), reminderTime.Minute(), 0, 0, s.location)
	
	return s.isTimeMatch(currentTime, targetTime, schedule.Expiration, eventName)
}

func (s *eventService) checkWeeklySchedule(currentTime time.Time, minutesBeforeEvent int64, startDate time.Time, schedule ScheduleSettings, eventName string) bool {
	dayMap := map[string]time.Weekday{
		"sunday": time.Sunday, "monday": time.Monday, "tuesday": time.Tuesday,
		"wednesday": time.Wednesday, "thursday": time.Thursday,
		"friday": time.Friday, "saturday": time.Saturday,
	}

	currentWeekday := currentTime.Weekday()
	log.Printf("📅 Event %s: current weekday = %s, checking days: %v",
		eventName, currentWeekday.String(), schedule.Day)

	for _, dayStr := range schedule.Day {
		if targetDay, ok := dayMap[strings.ToLower(dayStr)]; ok {
			if currentWeekday == targetDay {
				eventStartTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(),
					startDate.Hour(), startDate.Minute(), startDate.Second(), 0, s.location)
				reminderTime := eventStartTime.Add(-time.Duration(minutesBeforeEvent) * time.Minute)
				targetTime := time.Date(reminderTime.Year(), reminderTime.Month(), reminderTime.Day(),
					reminderTime.Hour(), reminderTime.Minute(), 0, 0, s.location)
				log.Printf("✅ Event %s: weekday matches %s, reminder time = %s",
					eventName, dayStr, targetTime.Format("2006-01-02 15:04:05"))
				return s.isTimeMatch(currentTime, targetTime, schedule.Expiration, eventName)
			}
		}
	}
	return false
}

func (s *eventService) checkMonthlySchedule(currentTime time.Time, minutesBeforeEvent int64, startDate time.Time, schedule ScheduleSettings, eventName string) bool {
	currentDay := currentTime.Day()
	log.Printf("📅 Event %s: current day = %d, checking days: %v",
		eventName, currentDay, schedule.Day)

	for _, dayStr := range schedule.Day {
		if day, err := strconv.Atoi(dayStr); err == nil && day >= 1 && day <= 31 {
			if currentDay == day {
				// Handle month-end edge cases (e.g., day 31 in February)
				eventStartTime := time.Date(currentTime.Year(), currentTime.Month(), day,
					startDate.Hour(), startDate.Minute(), startDate.Second(), 0, s.location)
				
				// Check if the day actually exists in current month
				if eventStartTime.Day() == day {
					reminderTime := eventStartTime.Add(-time.Duration(minutesBeforeEvent) * time.Minute)
					targetTime := time.Date(reminderTime.Year(), reminderTime.Month(), reminderTime.Day(),
						reminderTime.Hour(), reminderTime.Minute(), 0, 0, s.location)
					log.Printf("✅ Event %s: day matches %d, reminder time = %s",
						eventName, day, targetTime.Format("2006-01-02 15:04:05"))
					return s.isTimeMatch(currentTime, targetTime, schedule.Expiration, eventName)
				}
			}
		}
	}
	return false
}

func (s *eventService) checkYearlySchedule(currentTime time.Time, minutesBeforeEvent int64, startDate time.Time, schedule ScheduleSettings, eventName string) bool {
	eventStartTime := time.Date(currentTime.Year(), startDate.Month(), startDate.Day(),
		startDate.Hour(), startDate.Minute(), startDate.Second(), 0, s.location)
	
	reminderTime := eventStartTime.Add(-time.Duration(minutesBeforeEvent) * time.Minute)
	targetTime := time.Date(reminderTime.Year(), reminderTime.Month(), reminderTime.Day(),
		reminderTime.Hour(), reminderTime.Minute(), 0, 0, s.location)
	
	return s.isTimeMatch(currentTime, targetTime, schedule.Expiration, eventName)
}

func (s *eventService) isTimeMatch(currentTime, targetTime time.Time, expiration int, eventName string) bool {
	log.Printf("⏰ Event %s: comparing current=%s with target=%s",
		eventName, currentTime.Format("2006-01-02 15:04:05"), targetTime.Format("2006-01-02 15:04:05"))

	if currentTime.Equal(targetTime) {
		log.Printf("🎯 Event %s: exact time match!", eventName)
		return true
	}

	if expiration <= 1 {
		return false
	}

	log.Printf("🔄 Event %s: checking expiration, count=%d", eventName, expiration)
	for i := 1; i < expiration; i++ {
		nextTime := targetTime.Add(time.Duration(i*3) * time.Minute)
		log.Printf("Checking repeat %d: %s", i, nextTime.Format("2006-01-02 15:04:05"))
		if currentTime.Equal(nextTime) {
			log.Printf("🎯 Event %s: expiration time match (repeat %d)!", eventName, i)
			return true
		}
	}

	return false
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
	log.Printf("📨 Sending to %d tokens for user %s: %s", len(*tokens), event.UserID, event.EventName)

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
	if event.Reminders.Message != nil && *event.Reminders.Message != "" {
		return *event.Reminders.Message
	}
	return fmt.Sprintf("Nhắc nhở: %s sắp bắt đầu!", event.EventName)
}

func (s *eventService) GetAllEvents(ctx context.Context, userID string) ([]*Event, error) {
	
	if userID == "" {
		return nil, errors.New("user_id is required")
	}

	return s.eventRepository.FindAllEvents(ctx, userID)
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