package curriculumsvc

import (
	"testing"
	"time"
)

func ms(v int) *int { return &v }

func TestComputeRiskScore(t *testing.T) {
	tests := []struct {
		name      string
		events    []SecurityEvent
		wantScore int
		wantLevel string
	}{
		{"bersih", nil, 0, RiskLevelNormal},
		{"1x alt-tab sebentar", []SecurityEvent{{EventType: SecurityEventTabHidden, DurationMs: ms(2_000)}}, 4, RiskLevelNormal},
		{"2 fullscreen exit masih ditoleransi", []SecurityEvent{
			{EventType: SecurityEventFullscreenExit}, {EventType: SecurityEventFullscreenExit},
		}, 2, RiskLevelNormal},
		{"fullscreen exit ke-3 kena", []SecurityEvent{
			{EventType: SecurityEventFullscreenExit}, {EventType: SecurityEventFullscreenExit}, {EventType: SecurityEventFullscreenExit},
		}, 12, RiskLevelLowConfidence},
		{"durasi dicap 60 detik", []SecurityEvent{{EventType: SecurityEventWindowBlur, DurationMs: ms(10 * 60_000)}}, 14, RiskLevelLowConfidence},
		{"copy dicap 5 poin", func() []SecurityEvent {
			var ev []SecurityEvent
			for range 20 {
				ev = append(ev, SecurityEvent{EventType: SecurityEventCopyBlocked})
			}
			return ev
		}(), 5, RiskLevelNormal},
		{"sering buka tab lain lama", []SecurityEvent{
			{EventType: SecurityEventTabHidden, DurationMs: ms(30_000)},
			{EventType: SecurityEventTabHidden, DurationMs: ms(45_000)},
			{EventType: SecurityEventTabHidden, DurationMs: ms(20_000)},
		}, 31, RiskLevelReview},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeRiskScore(tt.events)
			if got != tt.wantScore {
				t.Fatalf("score = %d, want %d", got, tt.wantScore)
			}
			if lvl := riskLevelForScore(got); lvl != tt.wantLevel {
				t.Fatalf("level = %s, want %s", lvl, tt.wantLevel)
			}
		})
	}
}

func TestValidateSecurityEvents(t *testing.T) {
	now := time.Now()
	if err := validateSecurityEvents([]SecurityEvent{{EventType: "devtools_open", StartedAt: now}}); err == nil {
		t.Fatal("event type ngawur harusnya ditolak")
	}
	if err := validateSecurityEvents([]SecurityEvent{{EventType: SecurityEventTabHidden}}); err == nil {
		t.Fatal("startedAt kosong harusnya ditolak")
	}
	ev := []SecurityEvent{{EventType: SecurityEventTabHidden, StartedAt: now, DurationMs: ms(-5)}}
	if err := validateSecurityEvents(ev); err != nil || *ev[0].DurationMs != 0 {
		t.Fatalf("durasi negatif harusnya diclamp ke 0, got err=%v dur=%d", err, *ev[0].DurationMs)
	}
}
