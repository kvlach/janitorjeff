package discord

// Implement the core.Channel interface

type Channel struct {
	ChannelID string
}

func (ch *Channel) ID() string {
	return ch.ChannelID
}

func (ch *Channel) Name() string {
	return ch.ChannelID
}
