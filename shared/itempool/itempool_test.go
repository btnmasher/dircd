package itempool

import (
	"math/rand"
	"testing"
)

// Define a struct that implements ScrubbableItem
type mockItem struct {
	value int
	data  []int
}

func (i *mockItem) Scrub() {
	i.value = 0
	i.data = nil
}

func initItem() *mockItem {
	return &mockItem{
		data: make([]int, rand.Intn(100)),
	}
}

type testCase struct {
	num   int
	value int
}

func TestItemPool(t *testing.T) {
	cases := []testCase{
		{10, 0},
		{20, 0},
		{30, 0},
	}

	for _, tc := range cases {
		// Create a new pool and warm it up
		pool := New[*mockItem](100, initItem)
		pool.Warmup(tc.num)

		// Check if the items in the pool are as expected
		for i := 0; i < tc.num; i++ {
			item := pool.New()
			if item.value != tc.value {
				t.Errorf("Expected item value '%d', but got '%d'", tc.value, item.value)
			}

			// Set a random value for the item
			item.value = rand.Intn(100)

			// Recycle the item and check if the values have been reset
			pool.Recycle(item)
			if item.value != 0 {
				t.Errorf("Expected item value '0', but got '%d'", item.value)
			}
			if len(item.data) != 0 {
				t.Errorf("Expected item data slice to be empty, but it has %d elements", len(item.data))
			}
			if item.data != nil {
				t.Errorf("Expected item data slice to be nil, but it is not")
			}
		}
	}
}
