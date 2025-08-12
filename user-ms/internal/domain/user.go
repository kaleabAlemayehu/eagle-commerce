package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email     string             `json:"email" bson:"email" validate:"required,email"`
	Password  string             `json:"-" bson:"password" validate:"required,min=6"`
	FirstName string             `json:"first_name" bson:"first_name" validate:"required"`
	LastName  string             `json:"last_name" bson:"last_name" validate:"required"`
	Address   Address            `json:"address" bson:"address"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

type Address struct {
	Street  string `json:"street" bson:"street"`
	City    string `json:"city" bson:"city"`
	State   string `json:"state" bson:"state"`
	ZipCode string `json:"zip_code" bson:"zip_code"`
	Country string `json:"country" bson:"country"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, id string, user *User) (*User, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
}

type UserService interface {
	RegisterUser(ctx context.Context, user *User) (*User, string, error)
	GetUser(ctx context.Context, id string) (*User, error)
	UpdateUser(ctx context.Context, id string, user *User) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, limit, offset int) ([]*User, error)
	AuthenticateUser(ctx context.Context, email, password string) (*User, string, error)
}
