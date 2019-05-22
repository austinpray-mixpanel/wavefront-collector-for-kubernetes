package telegraf

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/influxdata/telegraf"
	gm "github.com/rcrowley/go-metrics"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
)

var filteredPoints gm.Counter

func init() {
	filteredPoints = gm.GetOrRegisterCounter("source.telegraf.points.filtered.count", gm.DefaultRegistry)
}

// Implements the telegraf Accumulator interface
type telegrafDataBatch struct {
	metrics.DataBatch
	source *telegrafPluginSource
}

func (t *telegrafDataBatch) preparePoints(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	var ts time.Time
	if len(timestamp) > 0 {
		ts = timestamp[0]
	} else {
		ts = time.Now()
	}

	for metric, v := range fields {
		var value float64
		var err error
		switch v.(type) {
		case string:
			continue
		default:
			value, err = getFloat(v)
			if err != nil {
				glog.Errorf("unsupported type: %v plugin: %s metric: %v value: %v. error: %v", reflect.TypeOf(v), t.source.name, metric, v, err)
				continue
			}
		}

		metricName := measurement + "." + metric
		metricName = strings.Replace(metricName, "_", ".", -1)
		if len(t.source.prefix) > 0 {
			metricName = t.source.prefix + "." + metricName
		}

		point := &metrics.MetricPoint{
			Metric:    metricName,
			Value:     value,
			Timestamp: ts.UnixNano() / 1000,
			Source:    t.source.source,
			Tags:      tags,
		}
		t.MetricPoints = t.filterAppend(t.MetricPoints, point)
	}
}

func (t *telegrafDataBatch) filterAppend(slice []*metrics.MetricPoint, point *metrics.MetricPoint) []*metrics.MetricPoint {
	if t.source.filters == nil || t.source.filters.Match(point.Metric, point.Tags) {
		return append(slice, point)
	}
	filteredPoints.Inc(1)
	glog.V(4).Infof("dropping metric: %s", point.Metric)
	return slice
}

// AddFields adds a metric to the accumulator with the given measurement
// name, fields, and tags (and timestamp). If a timestamp is not provided,
// then the accumulator sets it to "now".
func (t *telegrafDataBatch) AddFields(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	t.preparePoints(measurement, fields, tags, timestamp...)
}

// AddGauge is the same as AddFields, but will add the metric as a "Gauge" type
func (t *telegrafDataBatch) AddGauge(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	t.preparePoints(measurement, fields, tags, timestamp...)
}

// AddCounter is the same as AddFields, but will add the metric as a "Counter" type
func (t *telegrafDataBatch) AddCounter(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	t.preparePoints(measurement, fields, tags, timestamp...)
}

// AddSummary is the same as AddFields, but will add the metric as a "Summary" type
func (t *telegrafDataBatch) AddSummary(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	glog.Fatal("not supported")
}

// AddHistogram is the same as AddFields, but will add the metric as a "Histogram" type
func (t *telegrafDataBatch) AddHistogram(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	glog.Fatal("not supported")
}

// AddMetric adds an metric to the accumulator.
func (t *telegrafDataBatch) AddMetric(telegraf.Metric) {
	glog.Fatal("not supported")
}

// SetPrecision sets the timestamp rounding precision.  All metrics addeds
// added to the accumulator will have their timestamp rounded to the
// nearest multiple of precision.
func (t *telegrafDataBatch) SetPrecision(precision time.Duration) {
	glog.Fatal("not supported")
}

// Report an error.
func (t *telegrafDataBatch) AddError(err error) {
	glog.Fatal("not supported")
}

// Upgrade to a TrackingAccumulator with space for maxTracked metrics/batches.
func (t *telegrafDataBatch) WithTracking(maxTracked int) telegraf.TrackingAccumulator {
	glog.Fatal("not supported")
	return nil
}

var floatType = reflect.TypeOf(float64(0))

func getFloat(unk interface{}) (float64, error) {
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}
