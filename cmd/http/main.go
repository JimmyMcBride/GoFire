package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/JimmyMcBride/GoFire/config"
	"github.com/JimmyMcBride/GoFire/internal/storage"
	"github.com/JimmyMcBride/GoFire/internal/task"
	"github.com/JimmyMcBride/GoFire/pkg/shutdown"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

func main() {
	// setup exit code for graceful shutdown
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	// load config
	env, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("error: %v", err)
		exitCode = 1
		return
	}

	// run the server
	cleanup, err := run(env)

	// run the cleanup after the server is terminated
	defer cleanup()
	if err != nil {
		fmt.Printf("error: %v", err)
		exitCode = 1
		return
	}

	// ensure the server is shutdown gracefully & app runs
	shutdown.Gracefully()
}

func run(env config.EnvVars) (func(), error) {
	app, cleanup, err := buildServer(env)
	if err != nil {
		return nil, err
	}

	// start the server
	go func() {
		log.Fatal(app.Listen("0.0.0.0:" + env.PORT))
	}()

	// return a function to close the server and database
	return func() {
		cleanup()
		log.Fatal(app.Shutdown())
	}, nil
}

func buildServer(env config.EnvVars) (*fiber.App, func(), error) {
	// init the storage
	db, err := storage.BootstrapMongo(env.MongodbUri, env.MongodbName, 10*time.Second)
	if err != nil {
		return nil, nil, err
	}

	// Load and parse templates with custom functions
	engine := html.New("./views", ".gohtml")

	// create the fiber app
	app := fiber.New(
		fiber.Config{
			Views: engine,
		},
	)

	// add middleware
	app.Use(cors.New())
	app.Use(logger.New())

	// add health check
	app.Get(
		"/health", func(c *fiber.Ctx) error {
			return c.SendString("Healthy!")
		},
	)

	// create the task domain
	taskStore := task.NewTaskStorage(db)
	taskController := task.NewTaskController(taskStore)
	task.AddTaskRoutes(app, taskController)

	return app, func() {
		log.Fatal(storage.CloseMongo(db))
	}, nil
}
