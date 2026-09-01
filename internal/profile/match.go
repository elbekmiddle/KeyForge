package profile

import "strings"

// Match picks the best profile for the given window class / app name.
//
// Rule (doc section 23): a profile whose App field is a case-insensitive
// substring of class (or vice versa) wins; the longest matching App value
// wins ties, since it's the most specific. A profile with an empty App is
// treated as the fallback and only used if nothing else matches.
func Match(profiles []Profile, class string) (Profile, bool) {
	class = strings.ToLower(class)

	var (
		best     Profile
		bestLen  = -1
		fallback Profile
		hasFall  bool
	)

	for _, p := range profiles {
		app := strings.ToLower(strings.TrimSpace(p.App))

		if app == "" {
			if !hasFall {
				fallback = p
				hasFall = true
			}
			continue
		}

		if class != "" && (strings.Contains(class, app) || strings.Contains(app, class)) {
			if len(app) > bestLen {
				best = p
				bestLen = len(app)
			}
		}
	}

	if bestLen >= 0 {
		return best, true
	}

	return fallback, hasFall
}
