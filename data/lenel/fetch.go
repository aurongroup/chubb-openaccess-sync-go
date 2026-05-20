package lenel

import (
	"log"
	"openaccess-sync/client"
)

// fetchAndIndex fetches all instances of typeName from the API and returns them
// as an ordered slice, a map keyed by ID, and a map keyed by Name.
func fetchAndIndex[T any](
	cl *client.Client,
	typeName string,
	fromJSON func(map[string]any) (*T, error),
	getID func(*T) int32,
	getKey func(*T) string,
) ([]*T, map[int32]*T, map[string]*T, error) {
	items, err := cl.GetInstancesWithProgress(typeName, "")
	if err != nil {
		return nil, nil, nil, err
	}

	list := make([]*T, 0, len(items))
	byID := make(map[int32]*T, len(items))
	byKey := make(map[string]*T, len(items))

	for _, props := range items {
		item, err := fromJSON(props)
		if err != nil {
			log.Printf("skipping %s: %v", typeName, err)
			continue
		}
		list = append(list, item)
		byID[getID(item)] = item
		byKey[getKey(item)] = item
	}

	log.Printf("Retrieved %d %s records", len(list), typeName)
	return list, byID, byKey, nil
}
