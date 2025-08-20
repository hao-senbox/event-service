package event

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EventRepository interface {
	Create(ctx context.Context, event *Event) error
	FindEventActive(ctx context.Context) ([]*Event, error)
	FindAllEvents(ctx context.Context, userID string) ([]*Event, error)
	FindEventByID(ctx context.Context, eventID primitive.ObjectID) (*Event, error)
	UpdateEvent(ctx context.Context, event *Event, id primitive.ObjectID) error
	DeleteEvent(ctx context.Context, id primitive.ObjectID) error
}

type eventRepository struct {
	collection *mongo.Collection
}

func NewEventRepository(collection *mongo.Collection) EventRepository {
	_ = EnsureEventIndexes(context.Background(), collection)
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
		"is_send": true,
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

func (e *eventRepository) UpdateEvent(ctx context.Context, event *Event, id primitive.ObjectID) error {

	_, err := e.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": event})
	if err != nil {
		return err
	}
	return nil

}

func (e *eventRepository) DeleteEvent(ctx context.Context, id primitive.ObjectID) error {

	_, err := e.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	return nil

}

func EnsureEventIndexes(ctx context.Context, coll *mongo.Collection) error {

	models := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "is_send", Value: 1},
				{Key: "is_show", Value: 1},
				{Key: "end_date", Value: 1},
			},
			Options: options.Index().
				SetName("send_show_end"),
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().
				SetName("by_user_created"),
		},
	}
	_, err := coll.Indexes().CreateMany(ctx, models)
	return err
}
