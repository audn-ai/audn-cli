package policy

import "strings"

type Input struct {
    MinGrade        string // e.g., A/B/C/D/F
    FailOnSeverity  string // low/medium/high/critical
    VastGrade       string
    HighestSeverity string
}

type Result struct {
    Pass   bool
    Reason string
}

var severityRank = map[string]int{"low": 1, "medium": 2, "high": 3, "critical": 4}
var gradeRank = map[string]int{"A": 5, "B": 4, "C": 3, "D": 2, "F": 1}

func Evaluate(in Input) Result {
    minGrade := strings.ToUpper(strings.TrimSpace(in.MinGrade))
    if minGrade == "" { minGrade = "" }
    sevThresh := strings.ToLower(strings.TrimSpace(in.FailOnSeverity))
    if sevThresh == "" { sevThresh = "critical" }

    // Severity gate
    hs := strings.ToLower(strings.TrimSpace(in.HighestSeverity))
    if r(hs, severityRank) >= r(sevThresh, severityRank) {
        return Result{Pass: false, Reason: "a finding at or above the severity threshold was found"}
    }
    // Grade gate
    if minGrade != "" {
        vg := strings.ToUpper(strings.TrimSpace(in.VastGrade))
        if r(minGrade, gradeRank) > r(vg, gradeRank) {
            return Result{Pass: false, Reason: "VAST grade below required minimum"}
        }
    }
    return Result{Pass: true}
}

func r(key string, m map[string]int) int { if v, ok := m[key]; ok { return v }; return -1 }

