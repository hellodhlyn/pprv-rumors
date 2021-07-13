package main

import (
	"context"
	"time"

	"github.com/dstotijn/go-notion"
	"github.com/patrickmn/go-cache"
)

var notionClient *notion.Client
var cacheStore *cache.Cache

func GetBlockChildren(ctx context.Context, blockID string) (notion.BlockChildrenResponse, error) {
	result, err := readCache("block-"+blockID, func() (interface{}, error) {
		return notionClient.FindBlockChildrenByID(ctx, blockID, nil)
	})

	if err != nil {
		return notion.BlockChildrenResponse{}, err
	}
	return result.(notion.BlockChildrenResponse), nil
}

func GetDatabase(ctx context.Context, databaseID string) (notion.Database, error) {
	result, err := readCache("database-"+databaseID, func() (interface{}, error) {
		return notionClient.FindDatabaseByID(ctx, databaseID)
	})

	if err != nil {
		return notion.Database{}, err
	}
	return result.(notion.Database), nil
}

func QueryDatabase(ctx context.Context, databaseID string) (notion.DatabaseQueryResponse, error) {
	result, err := readCache("database-query-"+databaseID, func() (interface{}, error) {
		return notionClient.QueryDatabase(ctx, databaseID, &notion.DatabaseQuery{
			Sorts: []notion.DatabaseQuerySort{{Property: "Date", Direction: notion.SortDirDesc}}},
		)
	})

	if err != nil {
		return notion.DatabaseQueryResponse{}, err
	}
	return result.(notion.DatabaseQueryResponse), nil
}

func readCache(key string, fn func() (interface{}, error)) (interface{}, error) {
	if cacheResult, ok := cacheStore.Get(key); ok {
		return cacheResult, nil
	}

	result, err := fn()
	if err != nil {
		return nil, err
	}

	cacheStore.Set(key, result, 5*time.Minute)
	return result, nil
}
