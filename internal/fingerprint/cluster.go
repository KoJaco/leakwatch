package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"

	"github.com/KoJaco/leakwatch/internal/profile"
)

// FingerprintVersion is incremented when normalization rules change incompatibly.
const FingerprintVersion = 1

// ClusterSamples groups stack samples by normalized stack similarity.
func ClusterSamples(samples []profile.StackSample) []LeakCluster {
	if len(samples) == 0 {
		return nil
	}

	groups := make(map[string]*LeakCluster)
	for _, sample := range samples {
		if sample.Count <= 0 {
			continue
		}

		projected := ProjectStack(NormalizeStack(sample.Stack))
		if len(projected) == 0 {
			continue
		}

		key := stackKey(projected)
		cluster, ok := groups[key]
		if !ok {
			cluster = &LeakCluster{
				Fingerprint: Fingerprint{
					Key: key,
					ID:  fingerprintID(key),
				},
				Site: LeakSite{
					Location: leakLocation(projected),
					Stack:    append([]profile.Frame(nil), projected...),
				},
			}
			groups[key] = cluster
		}
		cluster.Count += sample.Count
	}
	if len(groups) == 0 {
		return nil
	}

	out := make([]LeakCluster, 0, len(groups))
	for _, c := range groups {
		out = append(out, *c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Fingerprint.Key < out[j].Fingerprint.Key
	})
	return out
}

func stackKey(stack []profile.Frame) string {
	var b strings.Builder
	for i, f := range stack {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(f.Function)
		b.WriteByte('|')
		b.WriteString(f.File)
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(f.Line))
	}
	return b.String()
}

func fingerprintID(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:8])
}
