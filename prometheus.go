package metrics

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"google.golang.org/protobuf/proto"
)

// -----------------------------------------------------------------------------

type ValueHandler func() float64

type VectorMetric []struct {
	Values  []string
	Handler ValueHandler
}

// -----------------------------------------------------------------------------

func (mws *Controller) CreateCounterWithCallback(name string, help string, handler ValueHandler) error {
	coll := prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		handler,
	)
	return mws.registry.Register(coll)
}

func (mws *Controller) CreateCounterVecWithCallback(
	name string, help string, variableLabels []string, subItems VectorMetric,
) error {
	desc := prometheus.NewDesc(
		prometheus.BuildFQName("", "", name),
		help,
		variableLabels,
		nil,
	)
	coll := &counterVecWithCallbackCollector{
		desc:    desc,
		metrics: make([]prometheus.Metric, 0),
	}
	for _, item := range subItems {
		if len(item.Values) != len(variableLabels) {
			return errors.New("invalid parameter")
		}

		m := &counterVecWithCallbackMetric{
			desc:    desc,
			handler: item.Handler,
		}
		m.self = m
		m.labelPairs = make([]*dto.LabelPair, 0)
		for idx, v := range item.Values {
			m.labelPairs = append(m.labelPairs, &dto.LabelPair{
				Name:  proto.String(variableLabels[idx]),
				Value: proto.String(v),
			})
		}

		coll.metrics = append(coll.metrics, m)
	}
	return mws.registry.Register(coll)
}

func (mws *Controller) CreateGaugeWithCallback(name string, help string, handler ValueHandler) error {
	coll := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		handler,
	)
	return mws.registry.Register(coll)
}

func (mws *Controller) CreateGaugeVecWithCallback(
	name string, help string, variableLabels []string, subItems VectorMetric,
) error {
	desc := prometheus.NewDesc(
		prometheus.BuildFQName("", "", name),
		help,
		variableLabels,
		nil,
	)
	coll := &gaugeVecWithCallbackCollector{
		desc:    desc,
		metrics: make([]prometheus.Metric, 0),
	}
	for _, item := range subItems {
		if len(item.Values) != len(variableLabels) {
			return errors.New("invalid parameter")
		}

		m := &gaugeVecWithCallbackMetric{
			desc:    desc,
			handler: item.Handler,
		}
		m.self = m
		m.labelPairs = make([]*dto.LabelPair, 0)
		for idx, v := range item.Values {
			m.labelPairs = append(m.labelPairs, &dto.LabelPair{
				Name:  proto.String(variableLabels[idx]),
				Value: proto.String(v),
			})
		}

		coll.metrics = append(coll.metrics, m)
	}
	return mws.registry.Register(coll)
}
