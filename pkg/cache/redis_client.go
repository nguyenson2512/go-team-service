package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"team-service/internal/entities"

	"github.com/redis/go-redis/v9"
)

type TeamCache interface {
	GetTeamMembers(ctx context.Context, teamID uint) ([]string, error)
	SetTeamMembers(ctx context.Context, teamID uint, members []string) error
	AddTeamMember(ctx context.Context, teamID uint, userID string) error
	RemoveTeamMember(ctx context.Context, teamID uint, userID string) error
}

type AssetCache interface {
	GetFolder(ctx context.Context, folderID uint) (*entities.Folder, error)
	SetFolder(ctx context.Context, folder *entities.Folder) error
	DeleteFolder(ctx context.Context, folderID uint) error

	GetNote(ctx context.Context, noteID uint) (*entities.Note, error)
	SetNote(ctx context.Context, note *entities.Note) error
	DeleteNote(ctx context.Context, noteID uint) error
}

// AccessControlCache interface for managing asset access control lists in Redis
type AccessControlCache interface {
	// SetAssetAccess sets the access type for a user on an asset
	SetAssetAccess(ctx context.Context, assetID string, userID string, accessType string) error

	// RemoveAssetAccess removes a user's access to an asset
	RemoveAssetAccess(ctx context.Context, assetID string, userID string) error

	// GetAssetAccess gets the access type for a user on an asset
	GetAssetAccess(ctx context.Context, assetID string, userID string) (string, error)

	// GetAssetACL gets all users and their access types for an asset
	GetAssetACL(ctx context.Context, assetID string) (map[string]string, error)
}

type redisTeamCache struct {
	client *redis.Client
}

type redisAssetCache struct {
	client *redis.Client
}

type redisAccessControlCache struct {
	client *redis.Client
}

func NewRedisTeamCache(addr string, password string, db int) TeamCache {
	cli := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	return &redisTeamCache{client: cli}
}

func NewRedisAssetCache(addr string, password string, db int) AssetCache {
	cli := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	return &redisAssetCache{client: cli}
}

func NewRedisAccessControlCache(addr string, password string, db int) AccessControlCache {
	cli := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	return &redisAccessControlCache{client: cli}
}

// Team cache methods
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

// Asset cache methods
func folderKey(folderID uint) string {
	return "folder:" + fmt.Sprint(folderID)
}

func noteKey(noteID uint) string {
	return "note:" + fmt.Sprint(noteID)
}

func (c *redisAssetCache) GetFolder(ctx context.Context, folderID uint) (*entities.Folder, error) {
	data, err := c.client.Get(ctx, folderKey(folderID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, err
	}

	var folder entities.Folder
	if err := json.Unmarshal([]byte(data), &folder); err != nil {
		return nil, err
	}

	return &folder, nil
}

func (c *redisAssetCache) SetFolder(ctx context.Context, folder *entities.Folder) error {
	data, err := json.Marshal(folder)
	if err != nil {
		return err
	}

	// Set with 1 hour expiration
	return c.client.Set(ctx, folderKey(folder.ID), data, time.Hour).Err()
}

func (c *redisAssetCache) DeleteFolder(ctx context.Context, folderID uint) error {
	return c.client.Del(ctx, folderKey(folderID)).Err()
}

func (c *redisAssetCache) GetNote(ctx context.Context, noteID uint) (*entities.Note, error) {
	data, err := c.client.Get(ctx, noteKey(noteID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, err
	}

	var note entities.Note
	if err := json.Unmarshal([]byte(data), &note); err != nil {
		return nil, err
	}

	return &note, nil
}

func (c *redisAssetCache) SetNote(ctx context.Context, note *entities.Note) error {
	data, err := json.Marshal(note)
	if err != nil {
		return err
	}

	// Set with 1 hour expiration
	return c.client.Set(ctx, noteKey(note.ID), data, time.Hour).Err()
}

func (c *redisAssetCache) DeleteNote(ctx context.Context, noteID uint) error {
	return c.client.Del(ctx, noteKey(noteID)).Err()
}

// Access control cache methods
func assetACLKey(assetID string) string {
	return "asset:" + assetID + ":acl"
}

func (c *redisAccessControlCache) SetAssetAccess(ctx context.Context, assetID string, userID string, accessType string) error {
	return c.client.HSet(ctx, assetACLKey(assetID), userID, accessType).Err()
}

func (c *redisAccessControlCache) RemoveAssetAccess(ctx context.Context, assetID string, userID string) error {
	return c.client.HDel(ctx, assetACLKey(assetID), userID).Err()
}

func (c *redisAccessControlCache) GetAssetAccess(ctx context.Context, assetID string, userID string) (string, error) {
	return c.client.HGet(ctx, assetACLKey(assetID), userID).Result()
}

func (c *redisAccessControlCache) GetAssetACL(ctx context.Context, assetID string) (map[string]string, error) {
	return c.client.HGetAll(ctx, assetACLKey(assetID)).Result()
}
