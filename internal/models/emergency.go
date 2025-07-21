package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Emergency struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	GroupID   primitive.ObjectID `bson:"group_id" json:"group_id"`
	UserID    string             `bson:"user_id" json:"user_id"`
	Type      string             `bson:"type" json:"type"`
	Message   *string            `bson:"message" json:"message"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdateAt  time.Time          `bson:"update_at" json:"update_at"`
}

type EmergencyLogs struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	EmergencyID primitive.ObjectID `bson:"emergency_id" json:"emergency_id"`
	GroupID     primitive.ObjectID `bson:"group_id" json:"group_id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	Status      string             `bson:"status" json:"status"`
	Type        string             `bson:"type" json:"type"`
	IsSender    bool               `bson:"is_sender" json:"is_sender"`
	Message     *string            `bson:"message" json:"message"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdateAt    time.Time          `bson:"update_at" json:"update_at"`
}

