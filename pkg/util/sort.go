package util

import (
	"sort"

	"github.com/kpiljoong/tome/pkg/model"
)

// SortEntriesByTimestampDesc sorts entries so that newest comes first.
func SortEntriesByTimestampDesc(entries []*model.JournalEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})
}

func SortJournalMapByTimestampDesc(entries map[string]*model.JournalEntry) []*model.JournalEntry {
	var sortedEntries []*model.JournalEntry
	for _, entry := range entries {
		sortedEntries = append(sortedEntries, entry)
	}
	sort.Slice(sortedEntries, func(i, j int) bool {
		return sortedEntries[i].Timestamp.After(sortedEntries[j].Timestamp)
	})
	return sortedEntries
}
