package inventory

import (
	"testing"

	"github.com/google/uuid"
)

func TestInventory(t *testing.T) {

	var (
		item1 = Item{
			Id:      uuid.New(),
			Name:    "Chelsea Boot",
			ForSale: true,
			Price:   30000,
		}

		pack1 = Pack{Type: item1, Size: 250}
		pack2 = Pack{Type: item1, Size: 500}
		pack3 = Pack{Type: item1, Size: 1000}
		pack4 = Pack{Type: item1, Size: 2000}
		pack5 = Pack{Type: item1, Size: 5000}
		ps    = NewPackSet()

		inv = Inventory{}
	)

	setup := func() {
		inv.data = ItemPackMap{}
		ps.Add(pack1)
		ps.Add(pack2)
		ps.Add(pack3)
		ps.Add(pack4)
		ps.Add(pack5)
		inv.data[item1.Id.String()] = *ps
	}

	t.Run("Inventory.ProcessOrder()", func(t *testing.T) {
		t.Run("Should return packs to fullfil an order", func(t *testing.T) {
			setup()

			result := inv.ProcessOrder(item1.Id.String(), 1)

			if len(result) != 1 {
				assertEqual(t, 1, len(result))
			}
			if result[pack1] != 1 {
				assertEqual(t, 1, result[pack1])
			}

			result = inv.ProcessOrder(item1.Id.String(), 12001)

			if len(result) != 3 {
				assertEqual(t, 3, len(result))
			}
			if result[pack5] != 2 {
				assertEqual(t, 2, result[pack5])
			}
			if result[pack5] != 2 {
				assertEqual(t, 2, result[pack5])
			}
			if result[pack4] != 1 {
				assertEqual(t, 1, result[pack4])
			}
			if result[pack1] != 1 {
				assertEqual(t, 1, result[pack1])
			}
		})
	})
}
