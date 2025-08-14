package event

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type EventRepository interface {
	Create(ctx context.Context, event *Event) error
	FindEventActive(ctx context.Context) ([]*Event, error)
	FindAllEvents(ctx context.Context, userID string) ([]*Event, error)
	FindEventByID(ctx context.Context, eventID primitive.ObjectID) (*Event, error)
}

type eventRepository struct {
	collection *mongo.Collection
}

func NewEventRepository(collection *mongo.Collection) EventRepository {
	return &eventRepository{
		collection: collection,
	}
}

func (e *eventRepository) Create(ctx context.Context, event *Event) error {

	_, err := e.collection.InsertOne(ctx, event)

	if err != nil {
		return err
	}

	return nil

}

func (e *eventRepository) FindEventActive(ctx context.Context) ([]*Event, error) {

	var events []*Event

	now := time.Now().UTC()
	fmt.Printf("🔍 Current UTC time: %s\n", now.Format("2006-01-02 15:04:05"))

	filter := bson.M{
		"active":                    true,
		"reminders.active_reminder": true,
		"end_date": bson.M{
			"$gte": now,
		},
	}

	cursor, err := e.collection.Find(ctx, bson.M{"active": true})
	if err != nil {
		return nil, err
	}

	var allActiveEvents []*Event
	err = cursor.All(ctx, &allActiveEvents)
	if err != nil {
		return nil, err
	}

	cursor, err = e.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	err = cursor.All(ctx, &events)
	if err != nil {
		return nil, err
	}

	fmt.Printf("✅ Found %d matching events\n", len(events))
	
	return events, nil
}

func (e *eventRepository) FindAllEvents(ctx context.Context, userID string) ([]*Event, error) {

	var events []*Event
	
	cursor, err := e.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &events)
	if err != nil {
		return nil, err
	}

	return events, nil

}

func (e *eventRepository) FindEventByID(ctx context.Context, eventID primitive.ObjectID) (*Event, error) {

	var event Event

	err := e.collection.FindOne(ctx, bson.M{"_id": eventID}).Decode(&event)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err	
	}

	return &event, err
	
}