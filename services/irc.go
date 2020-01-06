package services

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/privacylab/talek/libtalek"
)

// IRCSession represents the state of an IRC connection
type IRCSession struct {
	conn  net.Conn
	nick  string
	known []string
	chans []string
}

// NewIRCSession creates a new session object
func NewIRCSession(conn net.Conn, backend *libtalek.Client) *IRCSession {
	return &IRCSession{
		conn,
		"",
		[]string{},
		[]string{},
	}
}

// Watch waits for incoming connection lines and responds.
func (s *IRCSession) Watch() {
	reader := bufio.NewReader(s.conn)
	for {
		ba, _, err := reader.ReadLine()
		fmt.Println(string(ba))
		if err != nil {
			break
		}
		words := strings.Fields(string(ba))
		switch strings.ToLower(words[0]) {
		case "nick":
			s.nick = words[1]
			break
		case "user":
			s.user(words[1])
			break
		case "join":
			break
		case "part":
			break
		case "topic":
			break
		case "names":
			break
		case "list":
			break
		case "privmsg":
			break
		case "quit":
			s.conn.Close()
			break
		default:
			fmt.Fprintf(os.Stderr, "unknown IRC Command: %s\n", words[0])
		}
	}
}

func (s *IRCSession) respond(code int, msg []string) {
	s.conn.Write([]byte(fmt.Sprintf(":talirc %d %s\n", code, strings.Join(msg, " "))))
}

func (s *IRCSession) user(username string) {
	s.respond(1, []string{s.nick, ":welcome to talirc"})
}
