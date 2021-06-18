package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	var err error

	orderBook := NewOrderBook()
	msg := Message{}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		err = json.Unmarshal(scanner.Bytes(), &msg)
		if err != nil {
			fmt.Println("bad msg: ", err.Error())
		} else {
			switch msg.Kind {
			case KindTypeSnapshot:
				orderBook.Snapshot(msg)
				orderBook.Top()
			case KindTypeUpdate:
				if orderBook.Asks == nil || orderBook.Bids == nil {
					fmt.Println("update can't be before snapshot")
					continue
				}
				orderBook.Update(msg)
				orderBook.Top()
			default:
				fmt.Println("unknown command")
			}
		}
	}
}
