package model

import (
	"go.uber.org/zap"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const layout = "2006-01-02 15:04:05"

type Domain struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	DomainUrl   string             `bson:"domain" json:"domain"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Keywords    string             `bson:"keywords" json:"keywords"`
	Owner       string             `bson:"owner" json:"owner"`
	Expires     time.Time          `bson:"expires" json:"expires"`
	Created     time.Time          `bson:"created" json:"created"`
}

func convertTime(now time.Time) time.Time {
	// Format the time using the layout string
	formattedTime := now.Format(layout)
	// Convert the formatted time string to a time.Time value
	timeValue, err := time.Parse(layout, formattedTime)
	if err != nil {
		zap.L().Error(err.Error())
	}
	return timeValue
}
func (d *Domain) GetCreated() time.Time {
	return convertTime(d.Created)
}
func (d *Domain) GetExpires() time.Time {
	return convertTime(d.Expires)
}
func (d *Domain) GetDomainUrl() string {
	return d.DomainUrl
}
func (d *Domain) GetTitle() string {
	return d.Title
}
func (d *Domain) GetDescription() string {
	return d.Description
}
func (d *Domain) GetOwner() string {
	return d.Owner
}
func (d *Domain) GetKeywords() string {
	return d.Keywords
}
