package fingerprint

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"

	"github.com/KoJaco/leakwatch/internal/profile"
)

// ClusterGoroutines groups goroutines by normalized stack similarity.
func ClusterGoroutines(goroutines []profile.Goroutine) []LeakCluster {
	if len(goroutines) == 0 {
		return nil
	}

	groups := make(map[string]*LeakCluster)
	for _, g := range goroutines {
		projected := ProjectStack(NormalizeStack(g.Stack))
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
		cluster.Count++
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
	h := fnv.New32a()
	h.Write([]byte(key))
	return fmt.Sprintf("%08x", h.Sum32())
}
