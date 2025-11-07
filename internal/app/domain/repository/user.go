package repository

import (
	"github.com/7-solutions/backend-challenge/internal/app"
	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserRepository interface {
	CreateUser(c *fiber.Ctx, user *entity.User) error
	FindByEmail(c *fiber.Ctx, email string) (*entity.User, error)
	FindAll(c *fiber.Ctx) ([]*entity.User, error)
	FindAllPaginated(c *fiber.Ctx, filter bson.M, sortField string, sortOrder int, page, pageSize int) ([]*entity.User, bool, bool, error)
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

func (r *userRepository) FindByEmail(c *fiber.Ctx, email string) (*entity.User, error) {
	var user entity.User
	err := r.mongodb.FindOne(c.Context(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAll(c *fiber.Ctx) ([]*entity.User, error) {
	cursor, err := r.mongodb.Find(c.Context(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(c.Context())

	var users []*entity.User
	for cursor.Next(c.Context()) {
		var user entity.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) FindAllPaginated(
	c *fiber.Ctx,
	filter bson.M,
	sortField string,
	sortOrder int, // 1 for ascending, -1 for descending
	page, pageSize int,
) ([]*entity.User, bool, bool, error) {
	// Fetch pageSize + 1 to determine if there's a next page
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

	// Determine hasNext
	hasNext := len(users) > pageSize

	// Determine hasPrev
	hasPrev := page > 1

	return users, hasNext, hasPrev, nil
}

// GetCollection returns the MongoDB collection for probe queries
func (r *userRepository) GetCollection() *mongo.Collection {
	return r.mongodb
}

// GetFilter builds MongoDB filter from parameters
func (r *userRepository) GetFilter(filters map[string]interface{}) bson.M {
	filter := bson.M{}
	for key, value := range filters {
		if value != nil && value != "" {
			filter[key] = value
		}
	}
	return filter
}
