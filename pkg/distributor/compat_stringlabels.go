// SPDX-License-Identifier: AGPL-3.0-only

//go:build stringlabels

package distributor

import (
	"github.com/prometheus/prometheus/model/labels"

	"github.com/grafana/mimir/pkg/mimirpb"
)

func mergeActiveSeriesResponses(responses [][]*mimirpb.Metric) []labels.Labels {
	resultSet := make(map[labels.Labels]struct{})
	for _, resp := range responses {
		for _, series := range resp {
			lbls := mimirpb.FromLabelAdaptersToLabels(series.Labels)
			resultSet[lbls] = struct{}{}
		}
	}

	lbls := make([]labels.Labels, 0, len(resultSet))
	for v := range resultSet {
		lbls = append(lbls, v)
	}
	return lbls
}
