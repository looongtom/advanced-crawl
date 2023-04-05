package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	StatusEnable  = 1
	StatusDisable = 0
	StatusUpdated = 2
	StatusDone    = 3
)

type Domain struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	DomainUrl   string             `bson:"domain" json:"domain"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Keywords    string             `bson:"keywords" json:"keywords"`
	Owner       string             `bson:"owner" json:"owner"`
	Expires     time.Time          `bson:"expires" json:"expires"`
	Created     time.Time          `bson:"created" json:"created"`
	Status      int                `bson:"status" json:"status"`
}
