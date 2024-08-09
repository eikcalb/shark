package inventory

import (
	"errors"
	"log/slog"
	"testing"
)

const (
	MAX_UNBOUNDED_ITERATION_COUNT = 800

	NO_ERROR string = "(noerror)"
)

var (
	log *slog.Logger = slog.Default()

	ErrItemNotFound        = errors.New("item was not found")
	ErrIndexedItemNotFound = errors.New("indexed item was not found")
	ErrPackAlreadyExists   = errors.New("pack already exists in this set")
	ErrPackNotFound        = errors.New("pack was not found in this set")
)

func assertEqual[E interface{}, A interface{}](t *testing.T, expected E, actual A) {
	t.Fatalf("expected: %+v; got: %+v", expected, actual)
}
