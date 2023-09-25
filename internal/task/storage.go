package task

import (
	"context"
	"fmt"
	"math"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// how the task is stored in the database
type taskDB struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Completed   bool               `bson:"completed" json:"completed"`
}

type Pagination struct {
	Results      []taskDB `bson:"results" json:"results"`
	CurrentPage  int      `bson:"current-page" json:"current-page"`
	TotalPages   int      `bson:"total-pages" json:"total-pages"`
	TotalResults int      `bson:"total-results" json:"total-results"`
	Next         int      `bson:"next" json:"next"`
	Previous     int      `bson:"previous" json:"previous"`
}

type Storage struct {
	db *mongo.Database
}

type FilterOptions struct {
	Page      int
	PageSize  int
	Search    string
	Completed *bool
}

func NewTaskStorage(db *mongo.Database) *Storage {
	return &Storage{
		db: db,
	}
}

func NewFilterOptions() *FilterOptions {
	return &FilterOptions{
		Page:      1,
		PageSize:  10,
		Search:    "",
		Completed: nil,
	}
}

func (s *Storage) getAllTasks(ctx context.Context) ([]taskDB, error) {
	collection := s.db.Collection("todos")

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	tasks := make([]taskDB, 0)
	if err = cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *Storage) getPaginatedTasks(
	ctx context.Context,
	filterOptions *FilterOptions,
) (Pagination, error) {
	if filterOptions.Page < 1 {
		return Pagination{}, fmt.Errorf("page number must be greater than 0")
	}
	if filterOptions.PageSize < 1 {
		return Pagination{}, fmt.Errorf("page size must be greater than 0")
	}

	collection := s.db.Collection("todos")

	// Set up the filter for the query
	filter := bson.M{}
	if filterOptions.Search != "" {
		filter["$or"] = []bson.M{
			{"title": primitive.Regex{Pattern: filterOptions.Search, Options: "i"}},
			{"description": primitive.Regex{Pattern: filterOptions.Search, Options: "i"}},
		}
	}
	if filterOptions.Completed != nil {
		filter["completed"] = *filterOptions.Completed
	}

	// Get the total number of results
	totalResults, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return Pagination{}, err
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalResults) / float64(filterOptions.PageSize)))
	if totalPages == 0 {
		totalPages = 1
	}
	fmt.Printf("total pages: %d\n", totalPages)

	if filterOptions.Page > totalPages {
		return Pagination{}, fmt.Errorf("page number exceeds total pages")
	}

	// Calculate the skip number
	skip := (filterOptions.Page - 1) * filterOptions.PageSize

	// Set up find options with skip and limit
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(filterOptions.PageSize))
	findOptions.SetSort(bson.D{{Key: "_id", Value: -1}}) // reverse the order so that the newest task is on top

	// Retrieve the tasks for the current page
	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return Pagination{}, err
	}

	var tasks []taskDB
	if err = cursor.All(ctx, &tasks); err != nil {
		return Pagination{}, err
	}

	// Build the Pagination struct
	pagination := Pagination{
		CurrentPage:  filterOptions.Page,
		TotalPages:   totalPages,
		TotalResults: int(totalResults),
		Results:      tasks,
	}

	// Determine the Next and Previous page numbers
	if filterOptions.Page > 1 {
		pagination.Previous = filterOptions.Page - 1
	} else {
		pagination.Previous = 0 // Setting to 0 if there is no previous page
	}

	if filterOptions.Page < totalPages {
		pagination.Next = filterOptions.Page + 1
	} else {
		pagination.Next = 0 // Setting to 0 if there is no next page
	}

	return pagination, nil
}

func (s *Storage) getTaskByID(ctx context.Context, id string) (*taskDB, error) {
	collection := s.db.Collection("todos")

	// Convert the string ID to a MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Create a taskDB variable to hold the returned data
	var task taskDB

	// Find a single document by its ID
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&task)
	if err != nil {
		// You might get a mongo.ErrNoDocuments error here if the ID doesn't exist
		return nil, err
	}

	return &task, nil
}

func (s *Storage) createTask(title string, description string, completed bool, ctx context.Context) (
	string,
	error,
) {
	collection := s.db.Collection("todos")

	result, err := collection.InsertOne(ctx, bson.M{"title": title, "description": description, "completed": completed})
	if err != nil {
		return "", err
	}

	// convert the object id to a string
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (s *Storage) updateTaskCompletionStatus(ctx context.Context, id string, completed bool) error {
	collection := s.db.Collection("todos")

	// Convert the string ID to a MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Use the UpdateOne method to update the "completed" field of the task with the given ID
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"completed": completed}}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}
