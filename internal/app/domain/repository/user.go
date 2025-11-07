package repository

import (
	"errors"

	"github.com/7-solutions/backend-challenge/internal/app"
	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserRepository interface {
	CreateUser(c *fiber.Ctx, user *entity.User) error
	FindByEmail(c *fiber.Ctx, email string, user *entity.User) error
	FindAllPaginated(c *fiber.Ctx, filter bson.M, sortField string, sortOrder int, page, pageSize int) ([]*entity.User, bool, bool, error)
	FindByID(c *fiber.Ctx, filter bson.M, user *entity.User) error
	UpdateUser(c *fiber.Ctx, userID bson.ObjectID, currentVersion int, updates bson.M) error
	GetCollection() *mongo.Collection
	GetFilter(filters map[string]interface{}) bson.M
}

type userRepository struct {
	mongodb *mongo.Collection
}

func NewUserRepository(mongodb *mongo.Client) UserRepository {
	return &userRepository{
		mongodb: mongodb.Database(app.Config.MongoDB.Database).Collection(entity.User{}.CollectionName()),
	}
}

func (r *userRepository) CreateUser(c *fiber.Ctx, user *entity.User) error {
	_, err := r.mongodb.InsertOne(c.Context(), user)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) FindByEmail(c *fiber.Ctx, email string, user *entity.User) error {
	return r.mongodb.FindOne(c.Context(), bson.M{"email": email}).Decode(user)
}

func (r *userRepository) FindAllPaginated(
	c *fiber.Ctx,
	filter bson.M,
	sortField string,
	sortOrder int,
	page, pageSize int,
) ([]*entity.User, bool, bool, error) {
	limit := int64(pageSize + 1)
	skip := int64((page - 1) * pageSize)

	findOptions := options.Find().
		SetSort(bson.M{sortField: sortOrder}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.mongodb.Find(c.Context(), filter, findOptions)
	if err != nil {
		return nil, false, false, err
	}
	defer cursor.Close(c.Context())

	var users []*entity.User
	for cursor.Next(c.Context()) {
		var user entity.User
		if err := cursor.Decode(&user); err != nil {
			return nil, false, false, err
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, false, false, err
	}

	hasNext := len(users) > pageSize
	hasPrev := page > 1

	return users, hasNext, hasPrev, nil
}

func (r *userRepository) FindByID(c *fiber.Ctx, filter bson.M, user *entity.User) error {
	return r.mongodb.FindOne(c.Context(), filter).Decode(user)
}

func (r *userRepository) GetCollection() *mongo.Collection {
	return r.mongodb
}

func (r *userRepository) GetFilter(filters map[string]interface{}) bson.M {
	filter := bson.M{}
	for key, value := range filters {
		if value != nil && value != "" {
			filter[key] = value
		}
	}
	return filter
}

func (r *userRepository) UpdateUser(c *fiber.Ctx, userID bson.ObjectID, currentVersion int, updates bson.M) error {
	result, err := r.mongodb.UpdateOne(
		c.Context(),
		bson.M{
			"_id":     userID,
			"version": currentVersion,
		},
		bson.M{
			"$set": updates,
			"$inc": bson.M{"version": 1},
		},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("version conflict or user not found")
	}

	return nil
}
