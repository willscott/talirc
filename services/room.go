package services

import (
	"errors"
	"reflect"

	"github.com/privacylab/talek/libtalek"
)

type roomAction byte

const (
	participantAdd roomAction = iota
	participantRemove
	postMsg
	cancelWatch
)

type roomCommand struct {
	action roomAction
	msg    []byte
	*participant
}

type roomMsg struct {
	msg []byte
	err error
	*participant
}

type room struct {
	Participants []participant
	Log          chan roomMsg

	topic   *libtalek.Topic
	control chan roomCommand
}

func (r *room) postMsg(msg string) error {
	if r.control == nil {
		return errors.New("not watching room")
	}
	r.control <- roomCommand{postMsg, []byte(msg), nil}
	return nil
}

// watch a the room for new messages.
// this should be run a goroutine, and runs the main dynamo of room channels.
func (r *room) watch(client *libtalek.Client) error {
	if r.control != nil {
		return errors.New("already watching room")
	}
	r.control = make(chan roomCommand, 1)

	cases := make([]reflect.SelectCase, 1)
	cases[0] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(r.control)}
	for _, p := range r.Participants {
		incoming := client.Poll(p.Handle)
		p.selector = &reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(incoming)}
		cases = append(cases, *p.selector)
	}

	for {
		chosen, val, ok := reflect.Select(cases)
		if chosen == 0 {
			unwrapped, _ := val.Interface().(roomCommand)
			if unwrapped.action == cancelWatch {
				for _, p := range r.Participants {
					client.Done(p.Handle)
				}
				return nil
			} else if unwrapped.action == postMsg {
				err := client.Publish(r.topic, unwrapped.msg)
				if err != nil {
					r.Log <- roomMsg{nil, err, nil}
				}
			} else if unwrapped.action == participantAdd {
				r.Participants = append(r.Participants, *unwrapped.participant)
				incoming := client.Poll(unwrapped.participant.Handle)
				unwrapped.participant.selector = &reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(incoming)}
				cases = append(cases, *unwrapped.participant.selector)
			} else { // remove.
				for i := 0; i < len(cases); i++ {
					if cases[i].Chan == unwrapped.participant.selector.Chan {
						cases = append(cases[0:i], cases[i+1:]...)
					}
				}
				client.Done(unwrapped.participant.Handle)
			}
		} else {
			for _, p := range r.Participants {
				if *p.selector == cases[chosen] {
					r.onMsg(&p, val.Bytes())
					continue
				}
			}
			r.Log <- roomMsg{val.Bytes(), errors.New("room msg from unknown participant"), nil}
		}
		if !ok {
			return errors.New("unexpected channel close")
		}
	}
}

// stop watching a channel for new events.
func (r *room) unwatch() error {
	if r.control == nil {
		return errors.New("not watching room")
	}
	r.control <- roomCommand{cancelWatch, nil, nil}
	close(r.control)
	return nil
}

// Methods below this line are called internally within the room.
func (r *room) onMsg(from *participant, msg []byte) {
	// TOOD: is it a notification of membership change.
	r.Log <- roomMsg{msg, nil, from}
}
