/*
Package app implements a kernel that executes services.

This represents the root entry of the system and
is responsible for confirming dependencies as well as
configuring services that will be executed.
*/
package app

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"eikcalb.dev/shark/src/constants"
	"eikcalb.dev/shark/src/service"
)

var ErrApplication = errors.New("application experienced an error while running")

type Application struct {
	config *Config
	ctx    context.Context
	sm     *service.Services
}

func (app Application) setupServices() {
	slog.Info("Setting up application services")

	slog.Info("Service manager dsdsds", "kjk", app.sm)
	app.sm.Initialize(app.ctx)
}

// Run is called when the application should load and execute
// services.
//
// Each service run will receive a context function that can
// be used to exit the application from within the service.
func (app Application) Run() (err error) {
	slog.Info("Run application started")

	ctx, cancel := context.WithCancel(context.Background())
	// This is a guard to ensure there is no leak by informing
	// all cancellation channel listeners that the application
	// has exited.
	defer cancel()

	app.ctx = ctx
	app.setupServices()

	// This will handle any errors that occur due to the application
	// panicing.
	//
	// This could have been achieved via contexts, but with this method,
	// error caused in locations that do not have access to any context
	// will be caught.
	defer func() {
		// If an error occurs while the application is run, we want
		// to recover in order to report whatever might have occurred.
		slog.Info("Checking for panic to recover from")
		recovery := recover()
		if typedErr, ok := recovery.(error); ok {
			// The application panicked, and we have an error.
			slog.Error("Recovered from panic caused by the following error", "error", typedErr)
			err = errors.Join(ErrApplication, typedErr)
			return
		} else if recovery != nil {
			// We also handle situations where the type is not an error.
			// Because we are on golang 1.21.0, we should not be able to
			// panic with nil, so at this point we can assume there was a
			// panic.
			slog.Error("Recovered from panic without error", "data", recovery)
			err = ErrApplication
		}
	}()

	// Run registered services.
	ctx = context.WithValue(ctx, constants.CONTEXT_APPLICATION_VERSION_KEY, app.config.Version)
	ctx = context.WithValue(ctx, constants.CONTEXT_SERVICE_PORT_KEY, app.config.Port)
	go app.sm.Run(ctx)

	osSignalChannel := make(chan os.Signal, 1)
	signal.Notify(osSignalChannel, syscall.SIGINT, syscall.SIGTERM)

	<-osSignalChannel

	slog.Info("Run application ended")
	return err
}

func NewApplication(c *Config) *Application {
	slog.Info("Creating new Application instance with config", "config", c)
	slog.Default().With(c.Name, c.Version)

	app := &Application{}
	app.config = c
	app.sm = &service.Services{}

	slog.Info("Application instance created successfully")

	return app
}
