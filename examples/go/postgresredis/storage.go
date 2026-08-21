package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type visitStore struct {
	db *pgxpool.Pool
}

func (s visitStore) InsertVisit(ctx context.Context, user string, cacheCount int64) (time.Time, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var recordedAt time.Time
	err := s.db.QueryRow(
		queryCtx,
		`INSERT INTO visits (user_name, cache_count) VALUES ($1, $2) RETURNING created_at`,
		user,
		cacheCount,
	).Scan(&recordedAt)
	return recordedAt, err
}

func (s visitStore) AggregatedVisits(ctx context.Context) (map[string]int64, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	rows, err := s.db.Query(
		queryCtx,
		`SELECT user_name, COUNT(*) FROM visits GROUP BY user_name ORDER BY user_name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
		var user string
		var count int64
		if err := rows.Scan(&user, &count); err != nil {
			return nil, err
		}
		result[user] = count
	}
	return result, rows.Err()
}

type visitCache struct {
	client *redis.Client
}

func (c visitCache) IncrementVisit(ctx context.Context, user string) (int64, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	return c.client.Incr(queryCtx, cacheKey(user)).Result()
}

func (c visitCache) Snapshot(ctx context.Context) (map[string]int64, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	keys, err := c.client.Keys(queryCtx, "visits:*").Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64, len(keys))
	for _, key := range keys {
		value, err := c.client.Get(queryCtx, key).Result()
		if err != nil {
			return nil, err
		}
		count, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse redis count for %s: %w", key, err)
		}
		result[strings.TrimPrefix(key, "visits:")] = count
	}
	return result, nil
}

func cacheKey(user string) string {
	return "visits:" + user
}
