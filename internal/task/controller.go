package task

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	storage *Storage
}

func NewTaskController(storage *Storage) *Controller {
	return &Controller{
		storage: storage,
	}
}

type createTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (t *Controller) getBaseRoute(c *fiber.Ctx) error {
	filterOptions := NewFilterOptions()
	// get all tasks in a paginated fashion
	paginatedTasks, err := t.storage.getPaginatedTasks(c.Context(), filterOptions)
	if err != nil {
		// TODO: This needs to be an error page
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{
				"message": "Failed to get tasks",
			},
		)
	}

	return c.Render(
		"tasks", fiber.Map{
			"Title":        "Tasks",
			"Results":      paginatedTasks.Results,
			"CurrentPage":  paginatedTasks.CurrentPage,
			"TotalPages":   paginatedTasks.TotalPages,
			"TotalResults": paginatedTasks.TotalResults,
			"Next":         paginatedTasks.Next,
			"Previous":     paginatedTasks.Previous,
		}, "layouts/main",
	)
}

func (t *Controller) getTasks(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	fmt.Printf("Page number: %d\n", page)
	search := c.Query("search", "")
	completed := c.Query("completed")

	// Construct filter options using query parameters
	filterOptions := &FilterOptions{
		Page:     page,
		Search:   search,
		PageSize: 10,
	}

	if completed != "" {
		compBool, err := strconv.ParseBool(completed)
		if err == nil {
			filterOptions.Completed = &compBool
		}
	}

	// Get paginated tasks using the filter options
	paginatedTasks, err := t.storage.getPaginatedTasks(c.Context(), filterOptions)
	if err != nil {
		// TODO: Render an error page
		fmt.Printf("Error getting tasks: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{
				"message": "Failed to get tasks",
			},
		)
	}

	tmpl := template.Must(template.ParseFiles("views/tasks.gohtml"))

	buf := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(
		buf, "task-list-element", paginatedTasks,
	); err != nil {
		fmt.Printf("Error with template: %v\n", err)
		return c.Status(500).SendString(err.Error())
	}

	return c.Status(200).SendString(buf.String())
}

// Creates a new task and returns the updated list of tasks.
func (t *Controller) create(c *fiber.Ctx) error {
	time.Sleep(1 * time.Second)

	var req = createTaskRequest{Title: c.FormValue("title"), Description: c.FormValue("description")}

	// create the task
	_, err := t.storage.createTask(req.Title, req.Description, false, c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{
				"message": "Failed to create task",
			},
		)
	}

	filterOptions := NewFilterOptions()
	paginatedTasks, err := t.storage.getPaginatedTasks(c.Context(), filterOptions)

	tmpl := template.Must(template.ParseFiles("views/tasks.gohtml"))

	buf := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(buf, "task-list-element", paginatedTasks); err != nil {
		fmt.Printf("Error with template: %v\n", err)
		return c.Status(500).SendString(err.Error())
	}

	return c.Status(200).SendString(buf.String())
}

func (t *Controller) toggleCompletion(c *fiber.Ctx) error {
	// Parse task ID from the request
	taskID := c.Params("id")
	if taskID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"message": "Invalid task ID",
			},
		)
	}

	// Fetch the current task to get the current completion status
	task, err := t.storage.getTaskByID(c.Context(), taskID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{
				"message": "Failed to fetch the task",
			},
		)
	}

	// Toggle the completion status
	newStatus := !task.Completed

	// Update the task in the storage
	err = t.storage.updateTaskCompletionStatus(c.Context(), taskID, newStatus)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{
				"message": "Failed to update task status",
			},
		)
	}

	// Render and return the updated list of tasks
	page := c.QueryInt("page", 1)
	fmt.Printf("Page number: %d\n", page)
	search := c.Query("search", "")
	completed := c.Query("completed")

	// Construct filter options using query parameters
	filterOptions := &FilterOptions{
		Page:     page,
		Search:   search,
		PageSize: 10,
	}

	if completed != "" {
		compBool, err := strconv.ParseBool(completed)
		if err == nil {
			filterOptions.Completed = &compBool
		}
	}

	paginatedTasks, err := t.storage.getPaginatedTasks(c.Context(), filterOptions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{
				"message": "Failed to get tasks",
			},
		)
	}

	tmpl := template.Must(template.ParseFiles("views/tasks.gohtml"))

	buf := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(buf, "task-list-element", paginatedTasks); err != nil {
		fmt.Printf("Error with template: %v\n", err)
		return c.Status(500).SendString(err.Error())
	}

	return c.Status(200).SendString(buf.String())
}
