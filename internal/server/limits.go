package server

import "github.com/stockyard-dev/stockyard-cipher/internal/license"

type Limits struct {
	MaxProjects   int  // 0 = unlimited
	MaxSecrets    int  // total across all projects
	MaxTokens     int  // 0 = unlimited
	AuditLog      bool // Pro
	VersionHistory bool // Pro
	RetentionDays int
}

var freeLimits = Limits{
	MaxProjects:    2,
	MaxSecrets:     20,
	MaxTokens:      5,
	AuditLog:       false,
	VersionHistory: false,
	RetentionDays:  7,
}

var proLimits = Limits{
	MaxProjects:    0,
	MaxSecrets:     0,
	MaxTokens:      0,
	AuditLog:       true,
	VersionHistory: true,
	RetentionDays:  90,
}

func LimitsFor(info *license.Info) Limits {
	if info != nil && info.IsPro() {
		return proLimits
	}
	return freeLimits
}

func LimitReached(limit, current int) bool {
	if limit == 0 {
		return false
	}
	return current >= limit
}
