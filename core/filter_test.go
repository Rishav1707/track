package core

import (
	"testing"

	"github.com/mlange-42/track/util"
)

func TestFilters(t *testing.T) {
	tt := []struct {
		title   string
		filters []func(r *Record) bool
		records map[*Record]bool
	}{
		{
			title:   "no filters, all positive",
			filters: []func(r *Record) bool{},
			records: map[*Record]bool{
				{
					Project: "A",
				}: true,
				{
					Project: "B",
				}: true,
			},
		},
		{
			title: "filter by project names",
			filters: []func(r *Record) bool{
				FilterByProjects([]string{"A", "B"}),
			},
			records: map[*Record]bool{
				{
					Project: "A",
				}: true,
				{
					Project: "B",
				}: true,
				{
					Project: "C",
				}: false,
			},
		},
		{
			title: "filter by time",
			filters: []func(r *Record) bool{
				FilterByTime(util.DateTime(2000, 1, 1, 8, 0, 0), util.DateTime(2000, 1, 1, 20, 0, 0)),
			},
			records: map[*Record]bool{
				{
					Start: util.DateTime(2000, 1, 1, 1, 0, 0),
					End:   util.DateTime(2000, 1, 1, 2, 0, 0),
				}: false,
				{
					Start: util.DateTime(2000, 1, 1, 1, 0, 0),
					End:   util.DateTime(2000, 1, 1, 23, 0, 0),
				}: true,
				{
					Start: util.DateTime(2000, 1, 1, 1, 0, 0),
					End:   util.DateTime(2000, 1, 1, 9, 0, 0),
				}: true,
				{
					Start: util.DateTime(2000, 1, 1, 9, 0, 0),
					End:   util.DateTime(2000, 1, 1, 19, 0, 0),
				}: true,
				{
					Start: util.DateTime(2000, 1, 1, 19, 0, 0),
					End:   util.DateTime(2000, 1, 1, 22, 0, 0),
				}: true,
				{
					Start: util.DateTime(2000, 1, 1, 22, 0, 0),
					End:   util.DateTime(2000, 1, 1, 23, 0, 0),
				}: false,
			},
		},
		{
			title: "filter by any tags",
			filters: []func(r *Record) bool{
				FilterByTagsAny([]string{"A", "B"}),
			},
			records: map[*Record]bool{
				{
					Tags: []string{},
				}: false,
				{
					Tags: []string{"C", "D"},
				}: false,
				{
					Tags: []string{"A", "C"},
				}: true,
				{
					Tags: []string{"A", "B"},
				}: true,
				{
					Tags: []string{"A", "B", "C"},
				}: true,
			},
		},
		{
			title: "filter by all tags",
			filters: []func(r *Record) bool{
				FilterByTagsAll([]string{"A", "B"}),
			},
			records: map[*Record]bool{
				{
					Tags: []string{},
				}: false,
				{
					Tags: []string{"C", "D"},
				}: false,
				{
					Tags: []string{"A", "C"},
				}: false,
				{
					Tags: []string{"A", "B"},
				}: true,
				{
					Tags: []string{"A", "B", "C"},
				}: true,
			},
		},
	}

	for _, test := range tt {
		for rec, expOk := range test.records {
			ok := Filter(rec, FilterFunctions{test.filters, util.NoTime, util.NoTime})
			if ok != expOk {
				t.Fatalf("error when %s: expected %t, got %t for %v", test.title, expOk, ok, rec)
			}
		}
	}
}
