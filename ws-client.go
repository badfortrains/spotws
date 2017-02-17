package main

import (
	"errors"
	"github.com/badfortrains/spotcontrol"
	"net/rpc"
)

type Client struct {
	controllerMap *map[string]*spotcontrol.SpircController
	clientWsMap   *map[string]*rpc.Client
}

type IdentArgs struct {
	Ident string
	Token string
}

type SendVolumeArguments struct {
	Token  string
	Ident  string
	Volume int
}

type LoadTrackArguments struct {
	Token string
	Ident string
	Gids  []string
}

func (c *Client) getController(token string) (*spotcontrol.SpircController, error) {
	if controller, ok := (*c.controllerMap)[token]; ok {
		return controller, nil
	}
	return nil, errors.New("Authentication failed")
}

func (c *Client) SendHello(args *IdentArgs, _ *struct{}) error {
	controller, err := c.getController(args.Token)
	if err != nil {
		return err
	}
	return controller.SendHello()
}

func (c *Client) SendPlay(args *IdentArgs, _ *struct{}) error {
	controller, err := c.getController(args.Token)
	if err != nil {
		return err
	}
	return controller.SendPlay(args.Ident)
}

func (c *Client) SendPause(args *IdentArgs, _ *struct{}) error {
	controller, err := c.getController(args.Token)
	if err != nil {
		return err
	}
	return controller.SendPause(args.Ident)
}

func (c *Client) SendVolume(args *SendVolumeArguments, _ *struct{}) error {
	controller, err := c.getController(args.Token)
	if err != nil {
		return err
	}
	return controller.SendVolume(args.Ident, args.Volume)
}

func (c *Client) LoadTrack(args *LoadTrackArguments, _ *struct{}) error {
	controller, err := c.getController(args.Token)
	if err != nil {
		return err
	}
	return controller.LoadTrack(args.Ident, args.Gids)
}
