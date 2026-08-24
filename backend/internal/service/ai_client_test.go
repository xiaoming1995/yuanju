package service

import "testing"

func TestResolveStreamThinkingEnabled(t *testing.T) {
	tests := []struct {
		name                    string
		providerThinkingEnabled bool
		forceNoThink            bool
		want                    bool
	}{
		{
			name:                    "normal stream follows provider thinking",
			providerThinkingEnabled: true,
			forceNoThink:            false,
			want:                    true,
		},
		{
			name:                    "no-think stream overrides provider thinking",
			providerThinkingEnabled: true,
			forceNoThink:            true,
			want:                    false,
		},
		{
			name:                    "disabled provider remains disabled",
			providerThinkingEnabled: false,
			forceNoThink:            false,
			want:                    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveStreamThinkingEnabled(tt.providerThinkingEnabled, tt.forceNoThink); got != tt.want {
				t.Fatalf("resolveStreamThinkingEnabled(%v, %v) = %v, want %v",
					tt.providerThinkingEnabled, tt.forceNoThink, got, tt.want)
			}
		})
	}
}
