package types

import "testing"

func TestCalculateNetProfit(t *testing.T) {
	tests := []struct {
		name     string
		cost     float64
		sell     float64
		platform string
		wantNet  float64
		wantFee  float64
	}{
		{
			name:     "ebay fee",
			cost:     50,
			sell:     100,
			platform: "eBay",
			wantNet:  36.45,
			wantFee:  13.55,
		},
		{
			name:     "mercari fee",
			cost:     50,
			sell:     100,
			platform: "Mercari",
			wantNet:  40.00,
			wantFee:  10.00,
		},
		{
			name:     "unknown platform no fee",
			cost:     50,
			sell:     100,
			platform: "Other",
			wantNet:  50.00,
			wantFee:  0.00,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotNet, gotFee, _ := CalculateNetProfit(tc.cost, tc.sell, tc.platform)
			if gotNet != tc.wantNet {
				t.Fatalf("expected net %.2f, got %.2f", tc.wantNet, gotNet)
			}
			if gotFee != tc.wantFee {
				t.Fatalf("expected fee %.2f, got %.2f", tc.wantFee, gotFee)
			}
		})
	}
}

func TestFeeScheduleReturnsCopy(t *testing.T) {
	s := FeeSchedule()
	s["ebay"] = PlatformFee{}
	if FeeForPlatform("eBay").Percent == 0 {
		t.Fatal("expected fee schedule copy mutation not to affect defaults")
	}
}
