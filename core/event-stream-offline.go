package core

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
)

var (
	EventStreamOfflineHooks   = NewHooks[*EventStreamOffline](5)
	eventStreamOfflineChan    = make(chan *EventStreamOffline)
	eventStreamOfflineCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "jeff_event_stream_offline_total",
		Help: "Total number of received stream offline events.",
	}, []string{"frontend", "place"})
)

type EventStreamOffline struct {
	When     time.Time
	Here     Placer
	Frontend Frontender
}

func NewEventStreamOffline(when time.Time, here Placer, f Frontender) *EventStreamOffline {
	return &EventStreamOffline{
		When:     when,
		Here:     here,
		Frontend: f,
	}
}

func (soff *EventStreamOffline) Hooks() *Hooks[*EventStreamOffline] {
	return EventStreamOfflineHooks
}

func (soff *EventStreamOffline) Handler() {
	place, err := soff.Here.ScopeLogical()
	if err != nil {
		log.Error().Err(err)
		return
	}

	eventStreamOfflineCounter.With(prometheus.Labels{
		"frontend": soff.Frontend.Name(),
		"place":    strconv.FormatInt(place, 10),
	}).Inc()

	log.Debug().
		Str("when", soff.When.String()).
		Msg("received stream offline event")

	here, err := soff.Here.ScopeLogical()
	if err != nil {
		log.Error().Err(err).Msg("failed to get logical here")
		return
	}
	if err := DB.PlaceSet("stream_offline_actual", here, soff.When.UTC().Unix()); err != nil {
		log.Error().Err(err).Msg("failed save stream offline actual")
		return
	}
	EventStreamOfflineHooks.Run(soff)
}

func (soff *EventStreamOffline) Send() {
	eventStreamOfflineChan <- soff
}
