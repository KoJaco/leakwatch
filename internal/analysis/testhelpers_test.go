package analysis

import (
	"time"

	"github.com/KoJaco/leakwatch/internal/domain"
	"github.com/KoJaco/leakwatch/internal/fingerprint"
)

var testBaseTime = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func cluster(id string, count int) fingerprint.LeakCluster {
	return fingerprint.LeakCluster{
		Fingerprint: fingerprint.Fingerprint{ID: id, Key: id},
		Site: fingerprint.LeakSite{
			Location: fingerprint.Location{
				Function: "main.leak",
				File:     "main.go",
				Line:     42,
			},
		},
		Count: count,
	}
}

func obs(at time.Time, clusters ...fingerprint.LeakCluster) domain.Observation {
	return domain.Observation{
		CapturedAt: at,
		Clusters:   clusters,
	}
}

func minutes(base time.Time, offsets ...int) []time.Time {
	out := make([]time.Time, len(offsets))
	for i, off := range offsets {
		out[i] = base.Add(time.Duration(off) * time.Minute)
	}
	return out
}

func sequenceObs(id string, base time.Time, counts ...int) []domain.Observation {
	times := minutes(base, makeOffsets(len(counts))...)
	out := make([]domain.Observation, len(counts))
	for i, count := range counts {
		if count < 0 {
			out[i] = obs(times[i])
			continue
		}
		out[i] = obs(times[i], cluster(id, count))
	}
	return out
}

func makeOffsets(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i * 5
	}
	return out
}
