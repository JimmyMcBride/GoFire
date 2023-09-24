package film

import "go.mongodb.org/mongo-driver/mongo"

type Controller struct {
	db *mongo.Database
}

func NewFilmController(db *mongo.Database) *Controller {
	return &Controller{
		db: db,
	}
}
