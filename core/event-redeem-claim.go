package core

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
)

var (
	EventRedeemClaimHooks   = NewHooks[*EventRedeemClaim](5)
	eventRedeemClaimChan    = make(chan *EventRedeemClaim)
	eventRedeemClaimCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "jeff_event_redeem_claim_total",
		Help: "Total number of received redeem claim events.",
	}, []string{"frontend", "place"})
)

type EventRedeemClaim struct {
	ID       string
	Input    string
	When     time.Time
	Author   Personifier
	Here     Placer
	Frontend Frontender
}

func NewEventRedeemClaim(id, input string, when time.Time,
	author Personifier, here Placer, f Frontender) *EventRedeemClaim {

	return &EventRedeemClaim{
		ID:       id,
		Input:    input,
		When:     when,
		Author:   author,
		Here:     here,
		Frontend: f,
	}
}

func (rc *EventRedeemClaim) Hooks() *Hooks[*EventRedeemClaim] {
	return EventRedeemClaimHooks
}

func (rc *EventRedeemClaim) Handler() {
	place, err := rc.Here.ScopeLogical()
	if err != nil {
		log.Error().Err(err)
		return
	}

	eventRedeemClaimCounter.With(prometheus.Labels{
		"frontend": rc.Frontend.Name(),
		"place":    strconv.FormatInt(place, 10),
	}).Inc()

	log.Debug().
		Str("id", rc.ID).
		Str("input", rc.Input).
		Str("when", rc.When.String()).
		Msg("received redeem claim event")

	EventRedeemClaimHooks.Run(rc)
}

func (rc *EventRedeemClaim) Send() {
	eventRedeemClaimChan <- rc
}
