//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/janhq/jan-server/services/memory-tools/internal/configs"
)

func CreateApplication(cfg *configs.Config) (*Application, error) {
	wire.Build(newApplication)
	return nil, nil
}
