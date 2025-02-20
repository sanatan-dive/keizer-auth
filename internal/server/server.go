package server

import (
	"keizer-auth-api/internal/app"

	"github.com/gofiber/fiber/v2"
)

type FiberServer struct {
	*fiber.App
	container   *app.Container
	controllers *app.ServerControllers
}

func New() *FiberServer {
	container := app.GetContainer()
	controllers := app.GetControllers(container)

	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "keizer-auth-api",
			AppName:      "keizer-auth-api",
		}),
		container:   container,
		controllers: controllers,
	}

	return server
}

func (s *FiberServer) Shutdown() error {
	if err := s.App.Shutdown(); err != nil {
		return err
	}

	return s.container.Cleanup()
}
