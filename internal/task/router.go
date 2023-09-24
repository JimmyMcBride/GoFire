package task

import "github.com/gofiber/fiber/v2"

func AddTaskRoutes(app *fiber.App, controller *Controller) {
	tasks := app.Group("/tasks")

	// add middlewares here

	// add routes here
	tasks.Get("/", controller.getBaseRoute)
	tasks.Get("/get", controller.getTasks)
	tasks.Patch("/toggle/:id", controller.toggleCompletion)
	tasks.Post("/", controller.create)
}
