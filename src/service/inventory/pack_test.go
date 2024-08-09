package inventory

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

var (
	item1 = Item{
		Id:      uuid.New(),
		Name:    "Chelsea Boot",
		ForSale: true,
		Price:   30000,
	}

	item2 = Item{
		Id:      uuid.New(),
		Name:    "Trench Coat",
		ForSale: true,
		Price:   180000,
	}
)

func TestPackSet(t *testing.T) {
	t.Run("PackSet.Add()", func(t *testing.T) {
		t.Run("Should add packs to PackSet", func(t *testing.T) {
			ps := NewPackSet()

			pack1 := Pack{Type: item1, Size: 250}
			pack2 := Pack{Type: item1, Size: 280}
			pack3 := Pack{Type: item1, Size: 300}
			pack4 := Pack{Type: item1, Size: 2000}

			t.Logf("PackSet start size: %d", len(ps.values))

			ps.Add(pack1)
			ps.Add(pack2)
			ps.Add(pack3)
			ps.Add(pack4)

			actual := len(ps.values)
			if actual != 4 {
				assertEqual(t, 4, actual)
			}
		})

		t.Run("Should not add duplicate packs to PackSet", func(t *testing.T) {
			ps := NewPackSet()

			pack1 := Pack{Type: item1, Size: 250}

			ps.Add(pack1)
			err := ps.Add(pack1)

			if err == nil {
				assertEqual(t, ErrPackAlreadyExists, NO_ERROR)
			}
		})
	})

	t.Run("PackSet.Remove()", func(t *testing.T) {
		t.Run("Should remove packs to PackSet", func(t *testing.T) {
			ps := NewPackSet()

			pack1 := Pack{Type: item1, Size: 250}
			pack2 := Pack{Type: item1, Size: 280}

			ps.Add(pack1)
			ps.Add(pack2)

			length := len(ps.values)
			if length != 2 {
				assertEqual(t, 2, length)
			}

			err := ps.Remove(pack1)

			if err != nil {
				assertEqual(t, NO_ERROR, err)
			}

			actual := len(ps.values)
			if actual != 1 {
				assertEqual(t, 2, actual)
			}
		})

		t.Run("Should not remove pack if it does not exist in PackSet", func(t *testing.T) {
			ps := NewPackSet()

			pack1 := Pack{Type: item1, Size: 250}
			pack2 := Pack{Type: item1, Size: 280}

			ps.Add(pack1)

			actual := ps.Remove(pack2)

			if actual == nil {
				assertEqual(t, ErrPackNotFound, NO_ERROR)
			}
		})
	})
}

func TestItemPackMap(t *testing.T) {
	t.Run("ItemPackMap.Sort()", func(t *testing.T) {
		t.Run("Should sort packs for an item in ascending order", func(t *testing.T) {
			ps := NewPackSet()

			pack1 := Pack{Type: item1, Size: 250}
			pack2 := Pack{Type: item1, Size: 280}
			pack3 := Pack{Type: item1, Size: 300}
			pack4 := Pack{Type: item1, Size: 2000}

			ps.Add(pack4)
			ps.Add(pack2)
			ps.Add(pack3)
			ps.Add(pack1)

			im := ItemPackMap{}
			im[item1.Id.String()] = *ps

			t.Logf("ItemPackMap: %+v", im)

			err := im.Sort(item1.Id.String())
			if err != nil {
				assertEqual(t, NO_ERROR, err)
			}

			actual := ps.values[0]
			if actual != pack1 {
				assertEqual(t, pack1, actual)
			}
		})

		t.Run("Should not sort packs if item does not exist", func(t *testing.T) {
			ps := NewPackSet()

			pack1 := Pack{Type: item1, Size: 250}

			ps.Add(pack1)

			im := ItemPackMap{}
			im[item1.Id.String()] = *ps

			t.Logf("ItemPackMap: %+v", im)

			err := im.Sort(item2.Id.String())
			if !errors.Is(err, ErrIndexedItemNotFound) {
				assertEqual(t, ErrIndexedItemNotFound, err)
			}
		})
	})
}
