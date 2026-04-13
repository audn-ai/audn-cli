package policy

import "testing"

func TestEvaluateSeverityGate(t *testing.T) {
    in := Input{FailOnSeverity: "high", HighestSeverity: "critical"}
    if got := Evaluate(in); got.Pass {
        t.Fatalf("expected fail due to severity, got pass")
    }
}

func TestEvaluateGradeGate(t *testing.T) {
    in := Input{MinGrade: "B", VastGrade: "C", HighestSeverity: "low", FailOnSeverity: "critical"}
    if got := Evaluate(in); got.Pass {
        t.Fatalf("expected fail due to grade, got pass")
    }
}

func TestEvaluatePass(t *testing.T) {
    in := Input{MinGrade: "B", VastGrade: "A", HighestSeverity: "low", FailOnSeverity: "high"}
    if got := Evaluate(in); !got.Pass {
        t.Fatalf("expected pass, got fail: %v", got.Reason)
    }
}

