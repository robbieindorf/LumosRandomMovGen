package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	//"time"
)

const(
	baseApiPath  = ":8080/api"
	inventoryURL = baseApiPath + "/inventory"
	movesURL 	 = baseApiPath + "/moves"
)

// Args -l {Library IP} / -t {How long to run moves} / -m {Number of moves} / -p {Which partition to do moves for}
func main() {
	partition := "Auto Partition"
	var libraryIP string
	var numberOfMoves string

	argsWithoutProg := os.Args[1:]

	for i, v := range argsWithoutProg {
		switch v {
		case "-l":
			libraryIP = argsWithoutProg[i+1]
		case "-m":
			numberOfMoves = argsWithoutProg[i+1]
		case "-p":
			partition = argsWithoutProg[i+1]
		case "--help":
			fmt.Println("-l {Library IP}")
			fmt.Println("-m {Number of Moves}")
			fmt.Println("-p {Partition Name} (default: Auto Partition)")
		}
	}

	var wg sync.WaitGroup
	var parsedNumMoves int

	successfulMoves := 0
	failedMoves := 0

	successfulMovesChan := make(chan int, 1)
	failedMovesChan := make(chan int, 1)

	successfulMovesChan <- successfulMoves
	failedMovesChan <- failedMoves

	if numberOfMoves != "" {
		currentTime := time.Now().Format("01-02-06-15:04:05PM")
		logFileName := "moves_" + currentTime + ".log"
		file, err := os.OpenFile("./"+logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
		}
		defer file.Close()

		log.SetOutput(file)
		parsedNumMoves, err = strconv.Atoi(numberOfMoves)
		if err != nil {
			fmt.Println("Error parsing number of moves")
			return
		}
		rand.Seed(time.Now().UnixNano())
		log.Println("Library: " + libraryIP)
		log.Println("Partition: " + partition)
		log.Println("Move Count: " + numberOfMoves)
		for i := 0; i < parsedNumMoves; i++ {
			inventory, err := getInventory(libraryIP, partition)
			if err != nil {
				log.Println("Error: ", err)
				continue
			}

			move, err := generateMove(partition, inventory)

			moveID, err := sendMove(libraryIP, move)

			log.Println("#" + strconv.Itoa(i+1) + " - Move " + moveID + ": " + strconv.Itoa(move.Source) + " -> " + strconv.Itoa(move.Dest))
			go checkMoveStatus(&wg, libraryIP, moveID, successfulMovesChan, failedMovesChan)
			wg.Add(1)

		}
	} else {
		fmt.Println("Error not enough arguments / Use --help")
		return
	}

	wg.Wait()

	log.Println("Moves Count: ", parsedNumMoves, " / Successful: ", <- successfulMovesChan, " / Failed: ", <- failedMovesChan)
}
