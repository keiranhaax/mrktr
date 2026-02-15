package api

import "fmt"

// HTTPStatusError captures provider failures by status code.
type HTTPStatusError struct {
	Provider string
	Status   int
	Body     string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("%s status %d: %s", e.Provider, e.Status, e.Body)
}

func actionableProviderError(provider string, err error) string {
	statusErr, ok := err.(*HTTPStatusError)
	if !ok {
		return ""
	}

	switch {
	case statusErr.Status == 401 || statusErr.Status == 403:
		switch provider {
		case "Brave":
			return "Brave auth failed. Check BRAVE_API_KEY."
		case "Tavily":
			return "Tavily auth failed. Check TAVILY_API_KEY."
		case "Firecrawl":
			return "Firecrawl auth failed. Check FIRECRAWL_API_KEY."
		default:
			return fmt.Sprintf("%s auth failed. Check API key.", provider)
		}
	case statusErr.Status == 429:
		return fmt.Sprintf("%s rate limited. Try again in 60s.", provider)
	case statusErr.Status >= 500:
		return fmt.Sprintf("%s service error (%d). Try again shortly.", provider, statusErr.Status)
	default:
		return fmt.Sprintf("%s request failed (%d).", provider, statusErr.Status)
	}
}
