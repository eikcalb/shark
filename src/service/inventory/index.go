/*
Package inventory defines structures for representing items
and packs.

It also provides a service that can be used to create instances
of these structures and expose methods for manipulating these
instances.
*/
package inventory

import (
	"context"
	"log/slog"
	"sync"

	"eikcalb.dev/shark/src/constants"
	"eikcalb.dev/shark/src/store"
)

type InventoryOrder map[Pack]uint

type InventoryJSONFormat map[string][]Pack

type Inventory struct {
	data      ItemPackMap
	syncMutex sync.Mutex
	storage   *store.JSONFileStore[InventoryJSONFormat]
}

// getPacksForItemByID retrieves pascks for an Item with the ID
// specified in itemID.
func (i *Inventory) getPacksForItemByID(itemID string) (*PackSet, error) {
	for id, packSet := range i.data {
		if id == itemID {
			return &packSet, nil
		}
	}

	return nil, ErrItemNotFound
}

// lock sends data to syncChannel and helps to synchronize manipulation
// of the data collection.
func (i *Inventory) lock() {
	i.syncMutex.Lock()
}

// persist saves the inventory data for later retrieval.
func (i *Inventory) persist() {
	serializedData := i.serialize()
	log.Info("Attempting to save serialized data", "data", serializedData)

	err := i.storage.Save(*serializedData)
	if err != nil {
		log.Error("Failed to persist inventory data", "data", serializedData, "error", err)
	}

	log.Info("Successfully persisted inventory data")
}

// serialize converts inventory data to JSON format from an ItemPackMap and
// updates the inventory.
func (i *Inventory) serialize() *InventoryJSONFormat {
	log.Info("serialize data to JSON format start")

	i.lock()
	defer i.unLock()

	result := InventoryJSONFormat{}

	for id, packSet := range i.data {
		(result)[id] = packSet.getPacks()
	}

	log.Info("serialize data to JSON format end")
	return &result
}

// unLock receives data from syncChannel and helps to synchronize manipulation
// of the data collection.
func (i *Inventory) unLock() {
	i.syncMutex.Unlock()
}

// unserialize converts inventory data from JSON format to an ItemPackMap and
// updates the inventory.
func (i *Inventory) unserialize(jsonData *InventoryJSONFormat) {
	log.Info("unserialize data from JSON format start")

	i.lock()
	defer i.unLock()

	clear(i.data)
	for id, packs := range *jsonData {
		packSet := NewPackSet()
		for _, pack := range packs {
			log.Info("Add new pack to inventory", "pack", pack)
			err := packSet.Add(pack)
			if err != nil {
				log.Error("Failed to add new pack to inventory", "pack", pack, "error", err)
			}
		}
		packSet.Sort()
		i.data[id] = *packSet
	}

	log.Info("unserialize data from JSON format end")
}

func (i *Inventory) Initialize(ctx context.Context) error {
	// Logs should be scoped to make debugging easier.
	log = slog.Default().WithGroup("Inventory")

	log.Info("Initializing service")

	// The inventory data should be loaded into memory.
	jfs := store.JSONFileStore[InventoryJSONFormat]{Path: "storage.json"}
	jsonData, err := jfs.Load()
	if err != nil {
		// Failed to load config
		return err
	}

	i.storage = &jfs
	i.data = ItemPackMap{}
	// We have the JSON data, now we populate our application data.
	i.unserialize(jsonData)
	log.Info("serialized data from JSON", "data", i.data)

	return nil
}

// Run starts the inventory service.
func (i *Inventory) Run(ctx context.Context) error {
	var port uint16 = 44440
	// Fetch configuration from context and start running the service.
	rawPort := ctx.Value(constants.CONTEXT_SERVICE_PORT_KEY)
	port, ok := rawPort.(uint16)
	if !ok {
		log.Info("Failed to retrieve port from context, will use default")
	}

	i.startServer(ctx, port)

	return nil
}

// ProcessOrder accepts itemID as an identifier for an item in an order
// and count as the number of expected items in the order request. This
// method returns a map representing the packs that can be used in
// fulfilling the order with the Pack as its key and the frequency of each
// Pack as its value.
//
// We want to return the least number of packs o fulfill the order.
// Given the pack sizes, we will need to calculate how many packs will
// be required to fulfill count.
//
// Algorithm:
// In order to achieve this, we will iterate through the registered packs
// in ascending order. For example:
//   - If the order count for an item is 380.
//   - Assuming we have packs of 100, 200, 300 and 50.
//   - In order to efficiently handle the order, we would need to send
//     300 and 100.
//   - The best mental model to help understand is to imagine a truck that
//     helps deliver goods. When the item count is specified and no single
//     pack can represent the entire items, we use the nearest largest pack
//     to get most of the items and then get smaller packs so the truck is
//     not overloaded.
//   - We will find the smallest pack that is greater than or equal to the count.
//   - If the pack found is less than the required count, we will repeat this
//     logic and find the next pack greater than or equal to what is left.
//   - If there is no pack matching the criteria, then we will use the largest.
func (i *Inventory) ProcessOrder(itemID string, count int) InventoryOrder {
	var result = InventoryOrder{}
	log.Info("Process order start", "itemID", itemID, "count", count)

	// Get the PackSet referred to by itemID to fulfill the order.
	packs, err := i.getPacksForItemByID(itemID)
	if err != nil {
		log.Error("Failed to get item pack set", "itemID", itemID, "err", err)
		return result
	}

	// Now we will iterate the packs available and find the maximum pack to fulfill
	// this order. When found, we will store it's index and check for smaller packs
	// when count is less than the current pack.
	var currentPack *Pack
	currentIndex := 0
	currentCount := count

	packsSlice := packs.getPacks()
	length := len(packsSlice)
	for index, pack := range packsSlice {
		if int(pack.Size) >= currentCount {
			// Find first pack that is greater or equal to the current order count.
			currentPack = &pack
			currentIndex = index
			log.Info("Found matching pack", "itemID", itemID, "count", currentCount, "pack", pack)
			break
		}

		// If we are at the end of the loop, we use the largest pack.
		if index == length-1 {
			currentPack = &pack
			currentIndex = index
			log.Info("Using largest pack", "itemID", itemID, "count", currentCount, "pack", pack)
		}
	}

	// Prevents endless loops.
	iterCount := 0
	// We will continue decrementing the order count until it is less than or equal to 0.
	for currentCount > 0 && iterCount <= MAX_UNBOUNDED_ITERATION_COUNT {
		iterCount++
		log.Info("Updating order remaining count", "itemID", itemID, "count", currentCount, "size", currentPack.Size)

		// We need to check if the remaining orders fit into this pack or we need a
		// smaller pack. If the current count is less than this pack and we need to find a
		// smaller pack.
		if currentCount >= int(currentPack.Size) {
			// This pack can contain the items, so we subtract the current count.
			// Deduct the current pack size from the current count.
			currentCount -= int(currentPack.Size)
			result[*currentPack]++
			log.Info("Updated order remaining count", "itemID", itemID, "count", currentCount, "size", currentPack.Size)

			continue
		}

		// If the current count is less than the current pack siz then we need to check
		// for a smaller pack or we continue using the last pack available. The PackSet
		// returned we have is a copy of the actual PackSet making it
		// safe to trust that the index of packs will not change for out variable.
		if currentIndex > 0 {
			// If this is not the last pack, try using the next lower pack.
			currentIndex--
			currentPack = &packsSlice[currentIndex]
			log.Info("Changed pack", "itemID", itemID, "count", currentCount, "size", currentPack.Size)
			continue
		}
		// At this point, we are already at the smallest pack and we still have orders
		// to fulfill, so we use what we have.
		currentCount -= int(currentPack.Size)
		result[*currentPack]++
	}

	// Check that the item exists.
	log.Info("Process order success", "itemID", itemID, "count", currentCount, "result", result)
	return result
}
