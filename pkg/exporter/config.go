package exporter

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/blaxel-ai/kubernetes-event-exporter/pkg/kube"
	"github.com/blaxel-ai/kubernetes-event-exporter/pkg/sinks"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/rest"
)

const (
	DefaultCacheSize = 1024
	// DefaultFieldSelector is applied when Config.FieldSelector is empty.
	// These sources/reasons are by far the noisiest on real clusters and
	// are filtered out server-side by default. Users that genuinely want
	// those events can override by setting FieldSelector explicitly
	// (any non-empty value disables this default).
	DefaultFieldSelector = "source!=persistentvolume-controller,reason!=ProviderUpdateSuccess,reason!=ProviderDeleteSuccess"
	// DefaultReasonAllowlist drops any event whose reason does not match
	// this regex before it reaches the user's route engine. The set matches
	// the reasons consumed by the controlplane job analyzer. To disable,
	// set Config.ReasonAllowlist to ".*" (or any always-true regex).
	DefaultReasonAllowlist = "^(Scheduled|Pulling|Pulled|ProviderCreateSuccess|Created|Running|Succeeded|Completed|Failed|BackOff|FailedMount|FailedAttachVolume|Killing|Preempting|Terminated|OOMKilled|FailedScheduling|Error)$"
)

// Config allows configuration
type Config struct {
	// Route is the top route that the events will match
	// TODO: There is currently a tight coupling with route and config, but not with receiver config and sink so
	// TODO: I am not sure what to do here.
	LogLevel           string                    `yaml:"logLevel"`
	LogFormat          string                    `yaml:"logFormat"`
	ThrottlePeriod     int64                     `yaml:"throttlePeriod"`
	MaxEventAgeSeconds int64                     `yaml:"maxEventAgeSeconds"`
	ClusterName        string                    `yaml:"clusterName,omitempty"`
	Namespace          string                    `yaml:"namespace"`
	LeaderElection     kube.LeaderElectionConfig `yaml:"leaderElection"`
	Route              Route                     `yaml:"route"`
	Receivers          []sinks.ReceiverConfig    `yaml:"receivers"`
	KubeQPS            float32                   `yaml:"kubeQPS,omitempty"`
	KubeBurst          int                       `yaml:"kubeBurst,omitempty"`
	MetricsNamePrefix  string                    `yaml:"metricsNamePrefix,omitempty"`
	OmitLookup         bool                      `yaml:"omitLookup,omitempty"`
	CacheSize          int                       `yaml:"cacheSize,omitempty"`
	// FieldSelector is passed verbatim to the Events informer ListOptions, so the
	// apiserver filters events server-side and the exporter never receives them.
	// Example: "source!=persistentvolume-controller,source!=attachdetach-controller".
	// See https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/
	// When empty, defaults to DefaultFieldSelector. Set any non-empty value to opt out.
	FieldSelector string `yaml:"fieldSelector,omitempty"`
	// ReasonAllowlist is a regex evaluated against event.Reason before any route
	// processing. Events that do not match are dropped (and never reach a receiver).
	// When empty, defaults to DefaultReasonAllowlist. To disable, set to ".*".
	ReasonAllowlist string `yaml:"reasonAllowlist,omitempty"`
}

func (c *Config) SetDefaults() {
	if c.CacheSize == 0 {
		c.CacheSize = DefaultCacheSize
		log.Debug().Msg("setting config.cacheSize=1024 (default)")
	}

	if c.FieldSelector == "" {
		c.FieldSelector = DefaultFieldSelector
		log.Info().Str("fieldSelector", c.FieldSelector).Msg("setting config.fieldSelector (default)")
	}

	if c.ReasonAllowlist == "" {
		c.ReasonAllowlist = DefaultReasonAllowlist
		log.Info().Str("reasonAllowlist", c.ReasonAllowlist).Msg("setting config.reasonAllowlist (default)")
	}

	if c.KubeBurst == 0 {
		c.KubeBurst = rest.DefaultBurst
		log.Debug().Msg(fmt.Sprintf("setting config.kubeBurst=%d (default)", rest.DefaultBurst))
	}

	if c.KubeQPS == 0 {
		c.KubeQPS = rest.DefaultQPS
		log.Debug().Msg(fmt.Sprintf("setting config.kubeQPS=%.2f (default)", rest.DefaultQPS))
	}
}

func (c *Config) Validate() error {
	if err := c.validateDefaults(); err != nil {
		return err
	}
	if err := c.validateMetricsNamePrefix(); err != nil {
		return err
	}

	// No duplicate receivers
	// Receivers individually
	// Routers recursive
	return nil
}

func (c *Config) validateDefaults() error {
	if err := c.validateMaxEventAgeSeconds(); err != nil {
		return err
	}
	return nil
}

func (c *Config) validateMaxEventAgeSeconds() error {
	if c.ThrottlePeriod == 0 && c.MaxEventAgeSeconds == 0 {
		c.MaxEventAgeSeconds = 5
		log.Info().Msg("setting config.maxEventAgeSeconds=5 (default)")
	} else if c.ThrottlePeriod != 0 && c.MaxEventAgeSeconds != 0 {
		log.Error().Msg("cannot set both throttlePeriod (depricated) and MaxEventAgeSeconds")
		return errors.New("validateMaxEventAgeSeconds failed")
	} else if c.ThrottlePeriod != 0 {
		log_value := strconv.FormatInt(c.ThrottlePeriod, 10)
		log.Info().Msg("config.maxEventAgeSeconds=" + log_value)
		log.Warn().Msg("config.throttlePeriod is depricated, consider using config.maxEventAgeSeconds instead")
		c.MaxEventAgeSeconds = c.ThrottlePeriod
	} else {
		log_value := strconv.FormatInt(c.MaxEventAgeSeconds, 10)
		log.Info().Msg("config.maxEventAgeSeconds=" + log_value)
	}
	return nil
}

func (c *Config) validateMetricsNamePrefix() error {
	if c.MetricsNamePrefix != "" {
		// https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
		checkResult, err := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_:]*_$", c.MetricsNamePrefix)
		if err != nil {
			return err
		}
		if checkResult {
			log.Info().Msg("config.metricsNamePrefix='" + c.MetricsNamePrefix + "'")
		} else {
			log.Error().Msg("config.metricsNamePrefix should match the regex: ^[a-zA-Z][a-zA-Z0-9_:]*_$")
			return errors.New("validateMetricsNamePrefix failed")
		}
	} else {
		log.Warn().Msg("metrics name prefix is empty, setting config.metricsNamePrefix='event_exporter_' is recommended")
	}
	return nil
}
