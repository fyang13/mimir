package ingester

import (
	"context"
	"fmt"
	"github.com/grafana/dskit/user"
	"github.com/grafana/mimir/pkg/ingester/client"
	"github.com/grafana/mimir/pkg/mimirpb"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIngester_ActiveSeries(t *testing.T) {
	samples := []mimirpb.Sample{{TimestampMs: 1_000, Value: 1}}

	seriesWithLabelsOfSize := func(size, index int) mimirpb.PreallocTimeseries {
		// 24 bytes of static strings and slice overhead, the remaining bytes are used to
		// pad the value of the "lbl" label.
		require.Greater(t, size, 24, "minimum message size is 24 bytes")
		tpl := fmt.Sprintf("%%0%dd", size-24)
		return mimirpb.PreallocTimeseries{
			TimeSeries: &mimirpb.TimeSeries{
				Labels: mimirpb.FromLabelsToLabelAdapters(
					labels.FromStrings(labels.MetricName, "test", "lbl", fmt.Sprintf(tpl, index))),
				Samples: samples,
			},
		}
	}

	expectedMessageCount := 4
	totalSeriesSize := expectedMessageCount * activeSeriesMaxSizeBytes

	writeReq := &mimirpb.WriteRequest{Source: mimirpb.API}
	currentSize := 0
	for i := 0; currentSize < totalSeriesSize; i++ {
		s := seriesWithLabelsOfSize(1024, i)
		writeReq.Timeseries = append(writeReq.Timeseries, s)
		currentSize += s.Size()
	}

	// Write the series.
	ingesterClient := prepareHealthyIngester(t)
	ctx := user.InjectOrgID(context.Background(), userID)
	_, err := ingesterClient.Push(ctx, writeReq)
	require.NoError(t, err)

	// Get active series
	req := &client.ActiveSeriesRequest{
		MatchersSet: []*client.LabelMatchers{
			{Matchers: []*client.LabelMatcher{{Name: labels.MetricName, Value: "test", Type: client.EQUAL}}},
		},
	}

	server := &mockActiveSeriesServer{ctx: ctx}
	err = ingesterClient.ActiveSeries(req, server)
	require.NoError(t, err)

	// Check that all series were returned.
	returnedSeriesCount := 0
	for _, res := range server.responses {
		returnedSeriesCount += len(res.Metric)
	}
	assert.Equal(t, len(writeReq.Timeseries), returnedSeriesCount)

	// Check that we got the correct number of messages.
	assert.Equal(t, expectedMessageCount, len(server.responses))
}

type mockActiveSeriesServer struct {
	client.Ingester_ActiveSeriesServer
	responses []*client.ActiveSeriesResponse
	ctx       context.Context
}

func (s *mockActiveSeriesServer) Send(resp *client.ActiveSeriesResponse) error {
	s.responses = append(s.responses, resp)
	return nil
}

func (s *mockActiveSeriesServer) Context() context.Context {
	return s.ctx
}
