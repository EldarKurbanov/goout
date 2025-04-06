package telegram

import (
	"context"
	"fmt"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/gotd/td/telegram"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"goout/config"
	"goout/internal/integration/telegram/web"
)

type Client struct {
	client *gotgproto.Client

	webAuthClose func(context.Context) error
}

func NewClient(config *config.Config, errgroup *errgroup.Group, dialector gorm.Dialector) (*Client, error) {
	wa := web.GetWebAuth()
	// start web api
	waClose := web.Start(config, errgroup, wa)

	client, err := gotgproto.NewClient(
		// Get AppID from https://my.telegram.org/apps
		config.AppID,
		// Get ApiHash from https://my.telegram.org/apps
		config.APIHash,
		// ClientType, as we defined above
		gotgproto.ClientTypePhone(config.PhoneNumber),
		// Optional parameters of client
		&gotgproto.ClientOpts{

			// custom authenticator using web api
			AuthConversator: wa,
			Session:         sessionMaker.SqlSession(dialector),
			Device: &telegram.DeviceConfig{
				DeviceModel:    "web",
				SystemVersion:  "web",
				AppVersion:     "0.0.1",
				SystemLangCode: "en",
				LangCode:       "en",
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start client: %w", err)
	}

	return &Client{
		client:       client,
		webAuthClose: waClose,
	}, nil
}

func (c *Client) Stop(ctx context.Context) error {
	c.client.Stop()

	return c.webAuthClose(ctx)
}
