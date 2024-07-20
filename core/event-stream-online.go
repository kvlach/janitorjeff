package core

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
)

var (
	EventStreamOnlineHooks   = NewHooks[*EventStreamOnline](5)
	eventStreamOnlineChan    = make(chan *EventStreamOnline)
	eventStreamOnlineCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "jeff_event_stream_online_total",
		Help: "Total number of received stream online events.",
	}, []string{"frontend", "place"})
)

type EventStreamOnline struct {
	When     time.Time
	Here     Placer
	Frontend Frontender
}

func NewEventStreamOnline(when time.Time, here Placer, f Frontender) *EventStreamOnline {
	return &EventStreamOnline{
		When:     when,
		Here:     here,
		Frontend: f,
	}
}

func (son *EventStreamOnline) Hooks() *Hooks[*EventStreamOnline] {
	return EventStreamOnlineHooks
}

func (son *EventStreamOnline) Handler() {
	place, err := son.Here.ScopeLogical()
	if err != nil {
		log.Error().Err(err)
		return
	}

	eventStreamOnlineCounter.With(prometheus.Labels{
		"frontend": son.Frontend.Name(),
		"place":    strconv.FormatInt(place, 10),
	}).Inc()

	log.Debug().
		Str("when", son.When.String()).
		Msg("received stream online event")

	if err := son.save(); err != nil {
		log.Error().Err(err).Msg("failed to save stream online status")
		return
	}
	EventStreamOnlineHooks.Run(son)
}

// Save the timestamp of when the stream went online.
// Tries to filter shaky connections by giving a grace period of the stream going
// offline and online again (even multiple times), in which case the streams
// are viewed as one.
func (son *EventStreamOnline) save() error {
	// There are 2 kinds of values, actual and normalized. Actual keeps track of
	// online/offline events as they come in, without any filtering, normalized
	// makes sure that more than the specified grace period has passed between
	// the stream going down and up again.

	here, err := son.Here.ScopeLogical()
	if err != nil {
		return err
	}

	DB.Lock.Lock()
	defer DB.Lock.Unlock()

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()

	err = tx.PlaceSet("stream_online_actual", here, son.When.UTC().Unix())
	if err != nil {
		return err
	}

	offline, err := tx.PlaceGet("stream_offline_actual", here).Time()
	if err != nil {
		return err
	}

	grace, err := tx.PlaceGet("stream_grace", here).Duration()
	if err != nil {
		return err
	}

	diff := son.When.Sub(offline)
	if diff <= grace {
		log.Debug().
			Str("diff", diff.String()).
			Msg("stream online again within grace period")
		// in order to save the stream_online_actual value
		return tx.Commit()
	}

	offlinePrev, err := tx.PlaceGet("stream_offline_norm", here).Int64()
	if err != nil {
		return err
	}
	err = tx.PlaceSet("stream_offline_norm_prev", here, offlinePrev)
	if err != nil {
		return err
	}
	err = tx.PlaceSet("stream_offline_norm", here, offline.Unix())
	if err != nil {
		return err
	}
	err = tx.PlaceSet("stream_online_norm", here, son.When.UTC().Unix())
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (son *EventStreamOnline) Send() {
	eventStreamOnlineChan <- son
}
