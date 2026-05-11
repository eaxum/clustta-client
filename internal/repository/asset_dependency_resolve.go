package repository

import (
	"github.com/jmoiron/sqlx"
)

// ResolveBuildDependencies returns the rootAssetId followed by every asset id
// transitively reachable through asset_dependency and collection_dependency.
// Visiting a collection enqueues all of its assets and recurses into child
// collections. A visited set on both assets and collections prevents cycles
// and duplicate work. Trashed assets and collections are skipped.
func ResolveBuildDependencies(tx *sqlx.Tx, rootAssetId string) ([]string, error) {
	visitedAssets := map[string]bool{}
	visitedCollections := map[string]bool{}
	result := []string{}
	assetQueue := []string{}
	collectionQueue := []string{}

	enqueueAsset := func(id string) {
		if id == "" || visitedAssets[id] {
			return
		}
		visitedAssets[id] = true
		assetQueue = append(assetQueue, id)
		result = append(result, id)
	}

	enqueueCollection := func(id string) {
		if id == "" || visitedCollections[id] {
			return
		}
		visitedCollections[id] = true
		collectionQueue = append(collectionQueue, id)
	}

	enqueueAsset(rootAssetId)

	for len(assetQueue) > 0 || len(collectionQueue) > 0 {
		for len(assetQueue) > 0 {
			id := assetQueue[0]
			assetQueue = assetQueue[1:]

			depAssetIds := []string{}
			err := tx.Select(&depAssetIds, `
				SELECT ad.dependency_id
				FROM asset_dependency ad
				JOIN asset a ON a.id = ad.dependency_id
				WHERE ad.asset_id = ? AND a.trashed = 0
			`, id)
			if err != nil {
				return nil, err
			}
			for _, d := range depAssetIds {
				enqueueAsset(d)
			}

			depCollectionIds := []string{}
			err = tx.Select(&depCollectionIds, `
				SELECT cd.dependency_id
				FROM collection_dependency cd
				JOIN collection c ON c.id = cd.dependency_id
				WHERE cd.asset_id = ? AND c.trashed = 0
			`, id)
			if err != nil {
				return nil, err
			}
			for _, c := range depCollectionIds {
				enqueueCollection(c)
			}
		}

		for len(collectionQueue) > 0 {
			cid := collectionQueue[0]
			collectionQueue = collectionQueue[1:]

			assetsInCollection := []string{}
			err := tx.Select(&assetsInCollection,
				`SELECT id FROM asset WHERE collection_id = ? AND trashed = 0`, cid)
			if err != nil {
				return nil, err
			}
			for _, a := range assetsInCollection {
				enqueueAsset(a)
			}

			childCollections := []string{}
			err = tx.Select(&childCollections,
				`SELECT id FROM collection WHERE parent_id = ? AND trashed = 0`, cid)
			if err != nil {
				return nil, err
			}
			for _, ch := range childCollections {
				enqueueCollection(ch)
			}
		}
	}

	return result, nil
}
