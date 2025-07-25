package sinks

import (
	"context"
	"encoding/json"
	"log/syslog"

	"github.com/blaxel-ai/kubernetes-event-exporter/pkg/kube"
)

type SyslogConfig struct {
	Network string `yaml:"network"`
	Address string `yaml:"address"`
	Tag     string `yaml:"tag"`
}

type SyslogSink struct {
	sw *syslog.Writer
}

func NewSyslogSink(config *SyslogConfig) (Sink, error) {
	w, err := syslog.Dial(config.Network, config.Address, syslog.LOG_LOCAL0, config.Tag)
	if err != nil {
		return nil, err
	}
	return &SyslogSink{sw: w}, nil
}

func (w *SyslogSink) Close() {
	w.sw.Close()
}

func (w *SyslogSink) Send(ctx context.Context, ev *kube.EnhancedEvent) error {

	if b, err := json.Marshal(ev); err == nil {
		_, writeErr := w.sw.Write(b)

		if writeErr != nil {
			return writeErr
		}
	} else {
		return err
	}
	return nil
}
