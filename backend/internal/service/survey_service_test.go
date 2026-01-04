package service

import (
	"testing"

	"github.com/medical-app/backend/internal/entity"
)

func TestCalculateDAS28CRP(t *testing.T) {
	template := &entity.SurveyTemplate{Code: "DAS28_CRP"}

	tests := []struct {
		name         string
		responses    map[string]interface{}
		wantScore    float64
		wantCategory string
	}{
		{
			name: "remission",
			responses: map[string]interface{}{
				"tjc28": 0.0,
				"sjc28": 0.0,
				"crp":   1.0,
				"gh":    10.0,
			},
			// 0.56*sqrt(0) + 0.28*sqrt(0) + 0.36*ln(2) + 0.014*10 + 0.96 = 0.36*0.693 + 0.14 + 0.96 ≈ 1.35
			wantScore:    1.35,
			wantCategory: "remission",
		},
		{
			name: "moderate_activity",
			responses: map[string]interface{}{
				"tjc28": 2.0,
				"sjc28": 2.0,
				"crp":   5.0,
				"gh":    30.0,
			},
			// Computed: 3.21
			wantScore:    3.21,
			wantCategory: "moderate_activity",
		},
		{
			name: "moderate_activity_higher",
			responses: map[string]interface{}{
				"tjc28": 6.0,
				"sjc28": 4.0,
				"crp":   15.0,
				"gh":    50.0,
			},
			// Computed: 4.59
			wantScore:    4.59,
			wantCategory: "moderate_activity",
		},
		{
			name: "high_activity",
			responses: map[string]interface{}{
				"tjc28": 15.0,
				"sjc28": 12.0,
				"crp":   40.0,
				"gh":    80.0,
			},
			// Computed: 6.56
			wantScore:    6.56,
			wantCategory: "high_activity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call DAS28 calculator directly (bypasses JSON parsing)
			score, category, _, err := calculateDAS28CRP(template, tt.responses)
			if err != nil {
				t.Fatalf("calculateDAS28CRP() error = %v", err)
			}

			// Allow small floating point differences
			if diff := score - tt.wantScore; diff > 0.1 || diff < -0.1 {
				t.Errorf("calculateDAS28CRP() score = %v, want %v", score, tt.wantScore)
			}
			if category != tt.wantCategory {
				t.Errorf("calculateDAS28CRP() category = %v, want %v", category, tt.wantCategory)
			}
		})
	}
}

func TestCalculateBASDAI(t *testing.T) {
	template := &entity.SurveyTemplate{Code: "BASDAI"}

	tests := []struct {
		name         string
		responses    map[string]interface{}
		wantScore    float64
		wantCategory string
	}{
		{
			name: "low_activity",
			responses: map[string]interface{}{
				"q1": 2.0,
				"q2": 3.0,
				"q3": 2.0,
				"q4": 1.0,
				"q5": 3.0,
				"q6": 2.0,
			},
			// (2+3+2+1+(3+2)/2)/5 = (8+2.5)/5 = 2.1
			wantScore:    2.1,
			wantCategory: "low_activity",
		},
		{
			name: "high_activity",
			responses: map[string]interface{}{
				"q1": 7.0,
				"q2": 6.0,
				"q3": 8.0,
				"q4": 5.0,
				"q5": 7.0,
				"q6": 6.0,
			},
			// (7+6+8+5+(7+6)/2)/5 = (26+6.5)/5 = 6.5
			wantScore:    6.5,
			wantCategory: "high_activity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call BASDAI calculator directly (bypasses JSON parsing)
			score, category, _, err := calculateBASDAI(template, tt.responses)
			if err != nil {
				t.Fatalf("calculateBASDAI() error = %v", err)
			}

			if diff := score - tt.wantScore; diff > 0.1 || diff < -0.1 {
				t.Errorf("calculateBASDAI() score = %v, want %v", score, tt.wantScore)
			}
			if category != tt.wantCategory {
				t.Errorf("calculateBASDAI() category = %v, want %v", category, tt.wantCategory)
			}
		})
	}
}

func TestToFloat(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  float64
	}{
		{"float64", float64(3.14), 3.14},
		{"float32", float32(2.5), 2.5},
		{"int", int(42), 42.0},
		{"int64", int64(100), 100.0},
		{"nil", nil, 0.0},
		{"string", "not a number", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toFloat(tt.input)
			if got != tt.want {
				t.Errorf("toFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}
