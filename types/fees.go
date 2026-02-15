package types

import "strings"

// PlatformFee represents a platform selling fee.
type PlatformFee struct {
	Percent float64
	Flat    float64
}

var platformFees = map[string]PlatformFee{
	"ebay":     {Percent: 13.25, Flat: 0.30},
	"mercari":  {Percent: 10.0, Flat: 0.00},
	"amazon":   {Percent: 15.0, Flat: 0.00},
	"facebook": {Percent: 5.0, Flat: 0.00},
}

// FeeSchedule returns a copy of the default platform fee schedule.
func FeeSchedule() map[string]PlatformFee {
	out := make(map[string]PlatformFee, len(platformFees))
	for k, v := range platformFees {
		out[k] = v
	}
	return out
}

// FeeForPlatform returns fee settings for a platform.
func FeeForPlatform(platform string) PlatformFee {
	key := strings.ToLower(strings.TrimSpace(platform))
	if fee, ok := platformFees[key]; ok {
		return fee
	}
	return PlatformFee{}
}

// CalculateNetProfit computes net profit after platform fees.
func CalculateNetProfit(cost, sell float64, platform string) (net float64, fee float64, pct float64) {
	rule := FeeForPlatform(platform)
	fee = (sell * (rule.Percent / 100.0)) + rule.Flat
	if fee < 0 {
		fee = 0
	}
	net = sell - cost - fee
	if cost > 0 {
		pct = (net / cost) * 100
	}
	return net, fee, pct
}
