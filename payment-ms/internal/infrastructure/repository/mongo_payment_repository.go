package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/domain"
)

type MongoPaymentRepository struct {
	collection *mongo.Collection
}

func NewMongoPaymentRepository(db *mongo.Database) *MongoPaymentRepository {
	return &MongoPaymentRepository{
		collection: db.Collection("payments"),
	}
}

func (r *MongoPaymentRepository) Create(payment *domain.Payment) error {
	payment.ID = primitive.NewObjectID()
	payment.Status = domain.PaymentStatusPending
	payment.CreatedAt = time.Now()
	payment.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(context.Background(), payment)
	return err
}

func (r *MongoPaymentRepository) GetByID(id string) (*domain.Payment, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var payment domain.Payment
	err = r.collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&payment)
	if err != nil {
		return nil, err
	}

	return &payment, nil
}

func (r *MongoPaymentRepository) GetByOrderID(orderID string) (*domain.Payment, error) {
	var payment domain.Payment
	err := r.collection.FindOne(context.Background(), bson.M{"order_id": orderID}).Decode(&payment)
	if err != nil {
		return nil, err
	}

	return &payment, nil
}

func (r *MongoPaymentRepository) Update(id string, payment *domain.Payment) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	payment.UpdatedAt = time.Now()
	update := bson.M{"$set": payment}

	_, err = r.collection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
	return err
}

func (r *MongoPaymentRepository) UpdateStatus(id string, status domain.PaymentStatus) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
	return err
}

func (r *MongoPaymentRepository) List(limit, offset int) ([]*domain.Payment, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var payments []*domain.Payment
	for cursor.Next(context.Background()) {
		var payment domain.Payment
		if err := cursor.Decode(&payment); err != nil {
			return nil, err
		}
		payments = append(payments, &payment)
	}

	return payments, nil
}
