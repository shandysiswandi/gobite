package oauth

import "strings"

func buildProviders(custom []Provider) map[string]Provider {
	providers := map[string]Provider{}
	items := custom
	if len(items) == 0 {
		items = defaultProviders()
	}

	for _, item := range items {
		if item == nil {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(item.Name()))
		if name == "" {
			continue
		}
		providers[name] = item
	}

	return providers
}

func defaultProviders() []Provider {
	return []Provider{GoogleProvider{}, GitHubProvider{}}
}
