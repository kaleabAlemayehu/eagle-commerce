package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
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
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(id string, user *User) error
	Delete(id string) error
	List(limit, offset int) ([]*User, error)
}

type UserService interface {
	CreateUser(user *User) error
	GetUser(id string) (*User, error)
	UpdateUser(id string, user *User) error
	DeleteUser(id string) error
	ListUsers(limit, offset int) ([]*User, error)
	AuthenticateUser(email, password string) (*User, string, error)
}
