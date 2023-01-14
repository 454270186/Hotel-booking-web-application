package main

import (
	"fmt"
	"github.com/454270186/Hotel-booking-web-application/internal/config"
	"github.com/go-chi/chi/v5"
	"testing"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig

	h := routes(&app)
	switch varType := h.(type) {
	case *chi.Mux:
		// do nothing
	default:
		t.Error(fmt.Sprintf("type is not *chi.Mux, type is %T", varType))
	}

}
