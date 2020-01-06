package services

import (
	"errors"
	"fmt"
	"net"

	talekcommon "github.com/privacylab/talek/common"
	"github.com/privacylab/talek/libtalek"
)

// Serve returns an active Talirc listener.
func Serve(port int, talekConfig string) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	defer listener.Close()

	talekconf := libtalek.ClientConfigFromFile(talekConfig)
	if talekconf == nil {
		return errors.New("Could not load talek configuration file")
	}
	leader := talekcommon.NewFrontendRPC("rpc", talekconf.FrontendAddr)
	backend := libtalek.NewClient("talirc", *talekconf, leader)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		session := NewIRCSession(conn, backend)
		go session.Watch()
	}
}
