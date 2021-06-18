package main

import (
	"fmt"
	"sort"
	"sync"

	"github.com/cockroachdb/apd"
)

const (
	MAXOrderBookPositions = 100
)

type Message struct {
	Kind KindType          `json:"kind"`
	Asks map[string]string `json:"asks"`
	Bids map[string]string `json:"bids"`
}

type KindType string

const (
	KindTypeSnapshot KindType = "snapshot"
	KindTypeUpdate   KindType = "update"
)

type OrderBook struct {
	Asks []Order
	Bids []Order
}

type Order struct {
	Price    *apd.Decimal
	Quantity *apd.Decimal
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Asks: nil,
		Bids: nil,
	}
}

func (ob *OrderBook) Snapshot(msg Message) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	wg.Add(2)

	go func() {
		err := ob.setAsks(msg)
		if err != nil {
			fmt.Println("err setAsks", err)
			errCh <- err
		}

		wg.Done()
	}()

	go func() {
		err := ob.setBids(msg)
		if err != nil {
			fmt.Println("err setBids", err)
			errCh <- err
		}

		wg.Done()
	}()

	wg.Wait()
	close(errCh)
	// TODO: parse errCh and return errors

	return nil
}

func (ob *OrderBook) setAsks(msg Message) error {
	keys := make(sort.StringSlice, 0, len(msg.Asks))
	for k := range msg.Asks {
		keys = append(keys, k)
	}

	sort.Sort(keys)

	asks := make([]Order, 0, len(keys))
	for _, price := range keys {
		priceD, _, err := apd.NewFromString(price)
		if err != nil {
			return err
		}

		quantity := msg.Asks[price]
		quantityD, _, err := apd.NewFromString(quantity)
		if err != nil {
			return err
		}

		asks = append(asks, Order{
			Price:    priceD,
			Quantity: quantityD,
		})
	}

	ob.Asks = asks

	return nil
}

func (ob *OrderBook) setBids(msg Message) error {
	keys := make(sort.StringSlice, 0, len(msg.Bids))
	for k := range msg.Bids {
		keys = append(keys, k)
	}

	sort.Sort(sort.Reverse(keys[:]))
	if len(keys) > MAXOrderBookPositions {
		keys = keys[0 : MAXOrderBookPositions-1]
	}

	bids := make([]Order, 0, len(keys))
	for _, price := range keys {
		priceD, _, err := apd.NewFromString(price)
		if err != nil {
			return err
		}

		quantity := msg.Bids[price]
		quantityD, _, err := apd.NewFromString(quantity)
		if err != nil {
			return err
		}

		bids = append(bids, Order{
			Price:    priceD,
			Quantity: quantityD,
		})
	}

	ob.Bids = bids

	return nil
}

func (ob *OrderBook) Update(msg Message) {
	for price, quantity := range msg.Asks {
		// delete if quantity is zero
		if quantity == "0" {
			for index, order := range ob.Asks {
				if order.Price.String() == price {
					ob.Asks = RemoveIndex(ob.Asks, index)
				}
			}
		} else {
			found := false

			// update quantity if we have this price
			for index, order := range ob.Asks {
				if order.Price.String() == price {
					// TODO error check
					quantityD, _, _ := apd.NewFromString(quantity)
					ob.Asks[index].Quantity = quantityD
					found = true
				}
			}

			// add new Order
			if !found {
				// TODO we have an ordered slice, so we can easily insert a value to the right place.

				// if this is a new minimum price
				if price < ob.Asks[0].Price.String() {
					// TODO: add new order to the first place and move all other orders
					// if this price is bigger than TOP 100 - just skip it.
				} else if price > ob.Asks[len(ob.Asks)-1].Price.String() {
					// TODO: skip it
				} else {
					// TODO: let's add a new price to the right place in ordered slice.
				}
			}
		}
	}

	for price, quantity := range msg.Bids {
		// delete if quantity is zero
		if quantity == "0" {
			for index, order := range ob.Bids {
				if order.Price.String() == price {
					ob.Bids = RemoveIndex(ob.Bids, index)
				}
			}
		} else {
			found := false

			// update quantity if we have this price
			for index, order := range ob.Bids {
				if order.Price.String() == price {
					// TODO error check
					quantityD, _, _ := apd.NewFromString(quantity)
					ob.Bids[index].Quantity = quantityD
					found = true
				}
			}

			// add new Order
			if !found {
				// TODO we have an ordered slice, so we can easily insert a value to the right place.

				// if this is a new maximum price
				if price > ob.Bids[0].Price.String() {
					// TODO: add new order to the first place and move all other orders
					// if this price is less than TOP 100 - just skip it.
				} else if price < ob.Asks[len(ob.Bids)-1].Price.String() {
					// TODO: skip it
				} else {
					// TODO: let's add a new price to the right place in ordered slice.
				}
			}
		}
	}
}

func (ob *OrderBook) Top() {
	fmt.Printf(
		"Top ask: price=%s, quantity=%s; Top bid: price=%s, quantity=%s\n",
		ob.Asks[0].Price,
		ob.Asks[0].Quantity,
		ob.Bids[0].Price,
		ob.Bids[0].Quantity,
	)
}

func RemoveIndex(s []Order, index int) []Order {
	return append(s[:index], s[index+1:]...)
}
