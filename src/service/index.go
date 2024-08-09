/*
Package service defines tools to create services.
A service is a collection of business logic in an application.

A service can be seen as an endpoint in a microservice architecture.
*/
package service

import (
	"context"
	"log/slog"

	"eikcalb.dev/shark/src/service/inventory"
)

/*
Services is an interface that will expose methods used to manage
services within an application.
*/
type Service interface {
	Initialize(ctx context.Context) error
	Run(ctx context.Context) error
}

/*
Services is a collection of services that functions as a service
manager. Although not required, it helps organize responsibility
by managing services within the service package and exposing a
single structure to manage all services.
*/
type Services struct {
	services []Service
}

/*
Initialize is used to initialize services for the application.

TODO: Ideally, the services should be configurable via a config file.
*/
func (s *Services) Initialize(ctx context.Context) error {
	slog.Info("Service manager is initializing services")

	// Initialize services.
	inventory := &inventory.Inventory{}
	if err := inventory.Initialize(ctx); err != nil {
		slog.Error("Service manager encountered an error while inttializing inventory", "error", err)
		return err
	}

	// Store services.
	s.services = []Service{}
	s.services = append(s.services, inventory)

	return nil
}

func (s *Services) Run(ctx context.Context) error {
	slog.Info("Service manager is running services")

	for index, service := range s.services {
		go service.Run(ctx)
		slog.Info("Service started successfully", "sid", index)
	}

	// Listen for application exit.
	<-ctx.Done()

	slog.Info("Service manager has ended")
	return nil
}
