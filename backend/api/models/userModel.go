package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserModel struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	Email     string             `json:"email" bson:"email" validate:"required"`
	Password  string             `json:"password" bson:"password" validate:"required,min=5"`
	ImageUrl  string             `json:"imageUrl" bson:"imageUrl"`
	Bio       string             `json:"bio" bson:"bio"`
	Followers []string           `json:"followers" bson:"followers"`
	Following []string           `json:"following" bson:"following"`
}

// interfaces
type CreateUser struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type LoginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUser struct {
	Name     string `json:"name"`
	ImageUrl string `json:"imageUrl"`
	Bio      string `json:"bio"`
}
