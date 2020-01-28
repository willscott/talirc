package services

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/privacylab/talek/libtalek"
)

type roomAction byte

const (
	participantAdd roomAction = iota
	participantInvite
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
	// Participants is all other participants
	Participants []participant
	Log          chan roomMsg
	nickname     string

	// Outstanding invite to respond to.
	invite *libtalek.Topic
	// Pending sent invites waiting for response.
	pending []participant
	// My topic for publishing to.
	topic   *libtalek.Topic
	control chan roomCommand
}

func (r *room) UnmarshalJSON(data []byte) error {
	var vals map[string]json.RawMessage
	if err := json.Unmarshal(data, &vals); err != nil {
		return err
	}
	if t, ok := vals["topic"]; ok {
		// Full room serialized.
		var topic libtalek.Topic
		if err := json.Unmarshal(t, &topic); err != nil {
			return err
		}
		r.topic = &topic
		if h, ok := vals["pending"]; ok {
			if err := json.Unmarshal(h, &r.pending); err != nil {
				return err
			}
		}
		if n, ok := vals["nickname"]; ok {
			r.nickname = string(n)
		}
		if p, ok := vals["participants"]; ok {
			return json.Unmarshal(p, &r.Participants)
		}
		return nil
	} else if i, ok := vals["invite"]; ok {
		// Invite
		t, err := libtalek.NewTopic()
		if err != nil {
			return err
		}
		r.topic = t
		var invite libtalek.Topic
		if err := json.Unmarshal(i, &invite); err != nil {
			return err
		}
		r.invite = &invite

		if p, ok := vals["participants"]; ok {
			return json.Unmarshal(p, &r.Participants)
		}
		return nil
	}

	return &json.InvalidUnmarshalError{Type: reflect.TypeOf(r)}
}

func (r *room) MarshalJSON() ([]byte, error) {
	var m map[string]interface{}
	m["participants"] = r.Participants
	m["pending"] = r.pending
	m["nickname"] = r.nickname
	m["topic"] = r.topic
	return json.Marshal(m)
}

func (r *room) Invite() ([]byte, error) {
	invite, err := libtalek.NewTopic()
	if err != nil {
		return nil, err
	}
	initialTopic, err := invite.MarshalText()
	if err != nil {
		return nil, err
	}

	r.control <- roomCommand{participantInvite, initialTopic, nil}

	return initialTopic, nil
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

	cases := r.generateSelectors(client)

	for {
		chosen, val, ok := reflect.Select(cases)
		if chosen == 0 {
			unwrapped, _ := val.Interface().(roomCommand)
			if unwrapped.action == cancelWatch {
				for _, p := range r.Participants {
					client.Done(p.Handle)
				}
				for _, p := range r.pending {
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
				cases = r.generateSelectors(client)
			} else if unwrapped.action == participantInvite {
				t := new(libtalek.Topic)
				json.Unmarshal(unwrapped.msg, t)
				client.Publish(t, r.generateInvite(unwrapped.msg))

				r.pending = append(r.pending, participant{&t.Handle, "", nil})
				cases = r.generateSelectors(client)
			} else { // remove.
				client.Done(unwrapped.participant.Handle)
				r.removeParticipant(unwrapped.participant)
				cases = r.generateSelectors(client)
			}
		} else { // receive a message (either from current participant or pending)
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
	// TODO: is it a notification of membership change.
	r.Log <- roomMsg{msg, nil, from}
}

func (r *room) generateInvite(topic []byte) []byte {
	var m map[string]interface{}
	// todo: add self as a participant.
	m["participants"] = r.Participants
	m["invite"] = topic
	invite, err := json.Marshal(m)
	if err != nil {
		return nil
	}

	return invite
}

func (r *room) generateSelectors(client *libtalek.Client) []reflect.SelectCase {
	cases := make([]reflect.SelectCase, 1)
	cases[0] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(r.control)}
	for _, p := range r.Participants {
		if p.selector == nil {
			incoming := client.Poll(p.Handle)
			p.selector = &reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(incoming)}
		}
		cases = append(cases, *p.selector)
	}
	for _, p := range r.pending {
		if p.selector == nil {
			incoming := client.Poll(p.Handle)
			p.selector = &reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(incoming)}
		}
		cases = append(cases, *p.selector)
	}
	return cases
}

func (r *room) removeParticipant(item *participant) {
	for i, other := range r.Participants {
		if other == *item {
			r.Participants = append(r.Participants[:i], r.Participants[i+1:]...)
			return
		}
	}
}
