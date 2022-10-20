package test

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
)

func run(m *core.Message) (any, error, error) {
	switch m.Type {
	case core.Discord:
		return run_Discord(m)
	default:
		return run_Text(m)
	}
}

func run_Discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	embed := &dg.MessageEmbed{
		Description: run_Core(),
	}
	return embed, nil, nil
}

func run_Text(m *core.Message) (string, error, error) {
	return m.ReplyText(run_Core()), nil, nil
}

func run_Core() string {
	return "Test command!"
}
