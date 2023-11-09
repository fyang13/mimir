// SPDX-License-Identifier: AGPL-3.0-only

//go:build stringlabels

package distributor

import (
	"github.com/prometheus/prometheus/model/labels"

	"github.com/grafana/mimir/pkg/mimirpb"
)

// mergeActiveSeriesResponses takes a set of responses from different ingesters and merges them into a single set.
func mergeActiveSeriesResponses(responses [][]*mimirpb.Metric) []labels.Labels {
	// Build a unique set of labels to eliminate duplicates across responses.
	resultSet := make(map[labels.Labels]struct{})
	for _, resp := range responses {
		for _, series := range resp {
			lbls := mimirpb.FromLabelAdaptersToLabels(series.Labels)
			resultSet[lbls] = struct{}{}
		}
	}

	// Convert the set to a slice.
	lbls := make([]labels.Labels, 0, len(resultSet))
	for v := range resultSet {
		lbls = append(lbls, v)
	}

	return lbls
}
