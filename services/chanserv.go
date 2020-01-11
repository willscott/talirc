package services

import (
	"errors"

	"github.com/privacylab/talek/libtalek"
)

type chanserv struct {
	DB     map[string]*room
	client *libtalek.Client
}

func (c *chanserv) register(ch string) error {
	if _, ok := c.DB[ch]; ok {
		return errors.New("channel already exists")
	}
	c.DB[ch] = new(room)
	go c.DB[ch].watch(c.client)
	return nil
}

func (c *chanserv) invite(ch string) (string, error) {
	if _, ok := c.DB[ch]; !ok {
		return "", errors.New("channel not found")
	}

	return "", nil
}

func (c *chanserv) accept(token string) (string, error) {
	return "", nil
}
