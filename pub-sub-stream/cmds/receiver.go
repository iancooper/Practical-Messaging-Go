package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	dc "github.com/iancooper/Practical-Messaging-Go/pub-sub-stream/datachannel"
	"log"
	"os"
	"os/signal"
)

func main() {
	consumer := dc.NewConsumer("Pub-Sub-Stream-Biography",
		func(bytes []byte) (interface{}, error) {
			var bio = dc.Biography{}
			err := json.Unmarshal(bytes, &bio)
			return bio, err
		},
		func(message interface{}) error {
			bio := message.(dc.Biography)
			//add to MySQL
			db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/Lookup")
			if err != nil {
				log.Fatal(err)
			}
			defer db.Close()

			insertQuery := "INSERT INTO Biography (Id, Description) VALUES (?, ?)"
			_, err = db.Exec(insertQuery, bio.Id, bio.Description)
			if err != nil {
				log.Printf("Error inserting %s's biography: %v", bio.Id, err)
				return err
			} else {
				fmt.Printf("Successfully inserted %s's biography\n", bio.Id)
			}

			return nil
		},
	)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("Consumer exiting...")
		os.Exit(1)
	}()

	consumer.Receive()

}
