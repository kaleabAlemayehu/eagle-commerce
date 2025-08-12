package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/domain"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

type MongoProductRepository struct {
	collection *mongo.Collection
}

func NewMongoProductRepository(db *mongo.Database) *MongoProductRepository {
	return &MongoProductRepository{
		collection: db.Collection("products"),
	}
}

func (r *MongoProductRepository) Create(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	product.ID = primitive.NewObjectID()
	product.Active = true
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, product)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *MongoProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var product domain.Product
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&product)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	return &product, nil
}

func (r *MongoProductRepository) Update(ctx context.Context, id string, product *domain.Product) (*domain.Product, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	product.UpdatedAt = time.Now()
	update := bson.M{"$set": product}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedProduct domain.Product
	if err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID}, update, opts).Decode(&updatedProduct); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return &updatedProduct, nil
}

func (r *MongoProductRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{"$set": bson.M{"active": false, "updated_at": time.Now()}}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrProductNotFound
	}
	return err
}

func (r *MongoProductRepository) List(ctx context.Context, limit, offset int, category string) ([]*domain.Product, error) {
	filter := bson.M{"active": true}
	if category != "" {
		filter["category"] = category
	}

	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset))
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*domain.Product
	for cursor.Next(ctx) {
		var product domain.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return products, nil
}

func (r *MongoProductRepository) Search(ctx context.Context, query string, limit, offset int) ([]*domain.Product, error) {
	filter := bson.M{
		"active": true,
		"$or": []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset))
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*domain.Product
	for cursor.Next(ctx) {
		var product domain.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return products, nil
}

func (r *MongoProductRepository) UpdateStock(ctx context.Context, id string, quantity int) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$inc": bson.M{"stock": quantity},
		"$set": bson.M{"updated_at": time.Now()},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}
