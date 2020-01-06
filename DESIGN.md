Talek IRC
===

This design describes how an IRC messaging pattern is constructed using the talek single-writer-multiple-reader log primative.

* A channel #chan exists as a group messaging area, where all messages are designed to be read by other invited users in the room.
* A user constructs a channel by registering it against their proxy
    * `/msg chanserv register #chan`
    * The proxy creates a talek log for the channel, and stores the mapping of the readable name to the topic.
* Once a channel is registered, the user can request tracking and reciept of local messages via a standard `/join #chan` message.
* To invite a new participant
    * `/msg chanserv invite #chan`
    * chanserv will provide a token, which is passed to the user, who forwards it out of band to the new participant.
    * the new participant uses `/msg chanserv acceptinvite #chan <token>`.
* To accomplish the invitation user flow, the following actions must occur
    * The new participant must learn the public keys/handles for current channel participants.
    * The new participant must construct their topic handle for the channel.
    * current participants must learn of the new user and their topic handle.
* We accomplish these actions through the following steps:
    * When an invitation is generated, a new new topic log is generated.
    * The inviter writes a message contining the handles of all current participants in the channel
    * The private key of this log is then provided as the token.
    * The recipient reads the first message in the provided topic to learn the group when accepting invite.
    * The then generate their topic handle, and write the public key back to the topic that has been shared with them.
    * The inviter reads the new handle, and writes an add announcement on their channel handle with the new participant handle.
    * all other participants see this message and begin following the new handle.
