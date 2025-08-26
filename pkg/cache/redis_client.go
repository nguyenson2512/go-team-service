package cache

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

type TeamCache interface {
	GetTeamMembers(ctx context.Context, teamID uint) ([]string, error)
	SetTeamMembers(ctx context.Context, teamID uint, members []string) error
	AddTeamMember(ctx context.Context, teamID uint, userID string) error
	RemoveTeamMember(ctx context.Context, teamID uint, userID string) error
}

type redisTeamCache struct {
	client *redis.Client
}

func NewRedisTeamCache(addr string, password string, db int) TeamCache {
	cli := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	return &redisTeamCache{client: cli}
}

func teamMembersKey(teamID uint) string {
	return "team:" + fmt.Sprint(teamID) + ":members"
}

func (c *redisTeamCache) GetTeamMembers(ctx context.Context, teamID uint) ([]string, error) {
	return c.client.SMembers(ctx, teamMembersKey(teamID)).Result()
}

func (c *redisTeamCache) SetTeamMembers(ctx context.Context, teamID uint, members []string) error {
	if len(members) == 0 {
		return nil
	}

	return c.client.SAdd(ctx, teamMembersKey(teamID), members).Err()
}

func (c *redisTeamCache) AddTeamMember(ctx context.Context, teamID uint, userID string) error {
	return c.client.SAdd(ctx, teamMembersKey(teamID), userID).Err()
}

func (c *redisTeamCache) RemoveTeamMember(ctx context.Context, teamID uint, userID string) error {
	return c.client.SRem(ctx, teamMembersKey(teamID), userID).Err()
}