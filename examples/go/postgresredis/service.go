package main

import (
	"context"
	"fmt"
	"sort"
	"time"
)

type visitService struct {
	store visitStore
	cache visitCache
}

type visitResult struct {
	User       string    `json:"user"`
	CacheCount int64     `json:"cache_count"`
	RecordedAt time.Time `json:"recorded_at"`
}

type reportResult struct {
	Users []userVisits `json:"users"`
}

type userVisits struct {
	User       string `json:"user"`
	DBCount    int64  `json:"db_count"`
	CacheCount int64  `json:"cache_count"`
}

func (s visitService) RecordVisit(ctx context.Context, user string) (visitResult, error) {
	cacheCount, err := s.cache.IncrementVisit(ctx, user)
	if err != nil {
		return visitResult{}, fmt.Errorf("increment redis visit count: %w", err)
	}

	recordedAt, err := s.store.InsertVisit(ctx, user, cacheCount)
	if err != nil {
		return visitResult{}, fmt.Errorf("insert postgres visit: %w", err)
	}

	return visitResult{
		User:       user,
		CacheCount: cacheCount,
		RecordedAt: recordedAt,
	}, nil
}

func (s visitService) Report(ctx context.Context) (reportResult, error) {
	dbCounts, err := s.store.AggregatedVisits(ctx)
	if err != nil {
		return reportResult{}, fmt.Errorf("query postgres visit report: %w", err)
	}

	cacheCounts, err := s.cache.Snapshot(ctx)
	if err != nil {
		return reportResult{}, fmt.Errorf("query redis visit snapshot: %w", err)
	}

	usersByName := make(map[string]userVisits, len(dbCounts)+len(cacheCounts))
	for user, count := range dbCounts {
		entry := usersByName[user]
		entry.User = user
		entry.DBCount = count
		usersByName[user] = entry
	}
	for user, count := range cacheCounts {
		entry := usersByName[user]
		entry.User = user
		entry.CacheCount = count
		usersByName[user] = entry
	}

	users := make([]userVisits, 0, len(usersByName))
	for _, entry := range usersByName {
		users = append(users, entry)
	}
	sort.Slice(users, func(i, j int) bool {
		if users[i].DBCount == users[j].DBCount {
			return users[i].User < users[j].User
		}
		return users[i].DBCount > users[j].DBCount
	})

	return reportResult{Users: users}, nil
}
