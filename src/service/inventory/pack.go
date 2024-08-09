package inventory

import (
	"slices"

	"github.com/google/uuid"
)

// Item represents either a physical product or virtual goods.
type Item struct {
	// Id is a unique representation of this item.
	Id uuid.UUID `json:"id"`
	// Name is a non-unique representation of the item.
	Name string `json:"name"`

	// ForSale indicates whether this item is available for
	// sale or not.
	ForSale bool `json:"forSale"`
	// Price is an integer that shows the cost of this item.
	// It represents price multiplied by 100 to give an integer.
	Price uint32 `json:"price"`
}

// Pack is a collection of similar Items.
type Pack struct {
	// Type indicates the Item in contained within the pack.
	Type Item `json:"type"`
	// Size represents the number of items identified by Type
	// that are present in the pack.
	Size uint `json:"size"`
}

// PackSet is a collection of Pack structs that ensures its content
// is unique. This means no 2 packs will exist with the same Type
// and Size.
type PackSet struct {
	// keys represents the keys
	keys map[Pack]bool
	// values is the underlying data structure to store Pack instances.
	// The traditional map[Pack]bool interface, common in golang,
	// cannot be used because it introduces extra complexity in
	// sorting the Pack instances.
	values []Pack
}

// getPacks returns a copy of the values in this set. This is a slice of
// Pack pointers.
func (ps *PackSet) getPacks() []Pack {
	return ps.values
}

// Add accepts a Pack and throws an error if the pack already exists.
// Otherwise it inserts pack to the PackSet collection.
func (ps *PackSet) Add(pack Pack) error {
	if exists := ps.keys[pack]; exists {
		// The pack already exists. We will not keep quiet about this.
		// All values in a set should be unique.
		return ErrPackAlreadyExists
	}

	// We want to keep track of the keys to enforce uniqueness and also
	// track the slice containing packs.
	ps.keys[pack] = true
	ps.values = append(ps.values, pack)

	log.Info("Added new pack to set", "newEntry", pack, "set", ps.values)
	return nil
}

// Remove deletes pack from the collection.
func (ps *PackSet) Remove(pack Pack) error {
	// Because the packs can change order when sorted, we cannot store
	// the index of each pack. Otherwise we would be able to just store
	// the pack index as the value of Keys map. For now, we have to find
	// the pack's index and delete it.
	if packIndex := slices.Index(ps.values, pack); packIndex != -1 {
		delete(ps.keys, pack)
		ps.values = slices.Delete(ps.values, packIndex, packIndex+1)
		return nil
	} else {
		return ErrPackNotFound
	}
}

// Sort is used to sort pack by their Pack size.
func (ps *PackSet) Sort() {
	slices.SortStableFunc(ps.values, func(a, b Pack) int {
		return int(a.Size - b.Size)
	})
}

func NewPackSet() *PackSet {
	ps := &PackSet{}
	ps.keys = map[Pack]bool{}
	ps.values = []Pack{}
	return ps
}

// ItemPackMap is a container for associating each inventory Item
// with its packs.
type ItemPackMap map[string]PackSet

// Sort sorts packs for a given Item using the Pack size.
func (im ItemPackMap) Sort(itemID string) error {
	selectedItem, ok := im[itemID]
	if !ok {
		// Item not found in map.
		return ErrIndexedItemNotFound
	}
	selectedItem.Sort()

	return nil
}
