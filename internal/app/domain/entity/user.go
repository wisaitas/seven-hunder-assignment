package entity

type User struct {
	BaseEntity
	Name     string `bson:"name"`
	Email    string `bson:"email"`
	Password string `bson:"password"`
}

func (User) CollectionName() string {
	return "users"
}
