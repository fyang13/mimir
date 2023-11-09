// SPDX-License-Identifier: AGPL-3.0-only

//go:build !stringlabels

package distributor

import (
	"github.com/prometheus/prometheus/model/labels"

	"github.com/grafana/mimir/pkg/mimirpb"
)

// mergeActiveSeriesResponses takes a set of responses from different ingesters and merges them into a single set.
func mergeActiveSeriesResponses(responses [][]*mimirpb.Metric) []labels.Labels {
	type resultIndex struct {
		response int
		series   int
	}

	resultSet := make(map[uint64]resultIndex)
	for i, resp := range responses {
		for j, series := range resp {
			lbls := mimirpb.FromLabelAdaptersToLabels(series.Labels)
			resultSet[lbls.Hash()] = resultIndex{response: i, series: j}
		}
	}

	lbls := make([]labels.Labels, 0, len(resultSet))
	for _, idx := range resultSet {
		lbls = append(lbls, mimirpb.FromLabelAdaptersToLabels(responses[idx.response][idx.series].Labels))
	}
	return lbls
}
