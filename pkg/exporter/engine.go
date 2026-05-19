package exporter

import (
	"reflect"
	"regexp"

	"github.com/blaxel-ai/kubernetes-event-exporter/pkg/kube"
	"github.com/rs/zerolog/log"
)

// Engine is responsible for initializing the receivers from sinks
type Engine struct {
	Route           Route
	Registry        ReceiverRegistry
	ReasonAllowlist *regexp.Regexp
}

func NewEngine(config *Config, registry ReceiverRegistry) *Engine {
	for _, v := range config.Receivers {
		sink, err := v.GetSink()
		if err != nil {
			log.Fatal().Err(err).Str("name", v.Name).Msg("Cannot initialize sink")
		}

		log.Info().
			Str("name", v.Name).
			Str("type", reflect.TypeOf(sink).String()).
			Msg("Registering sink")

		registry.Register(v.Name, sink)
	}

	var allowlist *regexp.Regexp
	if config.ReasonAllowlist != "" {
		compiled, err := regexp.Compile(config.ReasonAllowlist)
		if err != nil {
			log.Fatal().Err(err).Str("reasonAllowlist", config.ReasonAllowlist).Msg("invalid reasonAllowlist regex")
		}
		allowlist = compiled
	}

	return &Engine{
		Route:           config.Route,
		Registry:        registry,
		ReasonAllowlist: allowlist,
	}
}

// OnEvent does not care whether event is add or update. Prior filtering should be done in the controller/watcher
func (e *Engine) OnEvent(event *kube.EnhancedEvent) {
	if e.ReasonAllowlist != nil && !e.ReasonAllowlist.MatchString(event.Reason) {
		log.Debug().Str("reason", event.Reason).Msg("event dropped by reasonAllowlist")
		return
	}
	e.Route.ProcessEvent(event, e.Registry)
}

// Stop stops all registered sinks
func (e *Engine) Stop() {
	log.Info().Msg("Closing sinks")
	e.Registry.Close()
	log.Info().Msg("All sinks closed")
}
