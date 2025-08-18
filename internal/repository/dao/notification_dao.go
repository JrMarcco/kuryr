package dao

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Notification struct {
	Id             primitive.ObjectID `json:"id" bson:"_id"`
	BizId          uint64             `json:"biz_id" bson: "biz_id"`
	BizKey         string             `json:"biz_key" bson: "biz_key"	`
	Receivers      string             `json:"receivers" bson: "receivers"`
	Channel        string             `json:"channel" bson: "channel"`
	TemplateId     uint64             `json:"template_id" bson: "template_id"`
	TemplateParams string             `json:"template_params" bson: "template_params"`
	SendStatus     string             `json:"send_status" bson: "send_status"`
	ScheduledStart int64              `json:"scheduled_start" bson: "scheduled_start"`
	ScheduledEnd   int64              `json:"scheduled_end" bson: "scheduled_end"`
	Version        int32              `json:"version" bson: "version"`
	CreatedAt      int64              `json:"created_at" bson: "created_at"`
	UpdatedAt      int64              `json:"updated_at" bson: "updated_at"`
}

func (n Notification) HexId() string {
	return n.Id.Hex()
}

type NotificationDao interface{}

type DefaultNotificationDao struct {
	client *mongo.Client
	db     *mongo.Database
}
