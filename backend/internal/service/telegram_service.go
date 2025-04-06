package service

import (
	"context"
	"errors"

	"golang.org/x/sync/errgroup"

	"goout/config"
	"goout/internal/integration/telegram"
	"goout/internal/repository"
)

type TelegramService struct {
	client     *telegram.Client
	repository *repository.TelegramRepository
}

func NewTelegramService(config *config.Config, errgroup *errgroup.Group) (*TelegramService, error) {
	repository, err := repository.NewTelegramRepository(config)
	if err != nil {
		return nil, err
	}

	client, err := telegram.NewClient(config, errgroup, repository.GetDialector())
	if err != nil {
		return nil, err
	}

	return &TelegramService{
		client:     client,
		repository: repository,
	}, nil
}

func (s *TelegramService) Stop(ctx context.Context) error {
	// Stop the client and repository if needed
	return errors.Join(
		s.client.Stop(ctx),
		s.repository.Stop(),
	)
}
