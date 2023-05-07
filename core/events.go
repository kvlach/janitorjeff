package core

import "github.com/rs/zerolog/log"

var (
	EventMessage      = make(chan *Message)
	EventMessageHooks = hooks[*Message]{}

	EventRedeemClaim      = make(chan string)
	EventRedeemClaimHooks = hooks[string]{}

	EventStreamOnline      = make(chan struct{})
	EventStreamOnlineHooks = hooks[struct{}]{}

	EventStreamOffline      = make(chan struct{})
	EventStreamOfflineHooks = hooks[struct{}]{}
)

func init() {
	go Events()
}

func Events() {
	for {
		select {
		case m := <-EventMessage:
			EventMessageHooks.Run(m)
			if _, err := m.CommandRun(); err != nil {
				log.Debug().Err(err).Send()
			}

		case redeemUUID := <-EventRedeemClaim:
			EventRedeemClaimHooks.Run(redeemUUID)

		case <-EventStreamOnline:
			EventStreamOnlineHooks.Run(struct{}{})

		case <-EventStreamOffline:
			EventStreamOfflineHooks.Run(struct{}{})
		}
	}
}
