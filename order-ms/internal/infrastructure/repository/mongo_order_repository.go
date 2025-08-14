package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/domain"
)

var (
	ErrOrderNotFound = errors.New("Order not found")
)

type MongoOrderRepository struct {
	collection *mongo.Collection
}

func NewMongoOrderRepository(db *mongo.Database) *MongoOrderRepository {
	return &MongoOrderRepository{
		collection: db.Collection("orders"),
	}
}

func (r *MongoOrderRepository) Create(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	order.ID = primitive.NewObjectID()
	order.Status = domain.OrderStatusPending
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	if _, err := r.collection.InsertOne(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (r *MongoOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var order domain.Order
	if err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&order); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	return &order, nil
}

func (r *MongoOrderRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Order, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*domain.Order
	for cursor.Next(ctx) {
		var order domain.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

func (r *MongoOrderRepository) Update(ctx context.Context, id string, order *domain.Order) (*domain.Order, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	order.UpdatedAt = time.Now()
	update := bson.M{"$set": order}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedOrder domain.Order
	if err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID}, update, opts).Decode(&updatedOrder); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return &updatedOrder, err
}

func (r *MongoOrderRepository) UpdateStatus(ctx context.Context, id string, currentStatus domain.OrderStatus, newStatus domain.OrderStatus) (*domain.Order, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     newStatus,
			"updated_at": time.Now(),
		},
	}
	filter := bson.M{
		"_id":    objectID,
		"status": currentStatus,
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedOrder domain.Order
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedOrder); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return &updatedOrder, err
}

func (r *MongoOrderRepository) List(ctx context.Context, limit, offset int) ([]*domain.Order, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*domain.Order
	for cursor.Next(ctx) {
		var order domain.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, nil
}
