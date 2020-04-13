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
	var timeLimit string

	argsWithoutProg := os.Args[1:]

	for i, v := range argsWithoutProg {
		switch v {
		case "-l":
			libraryIP = argsWithoutProg[i+1]
		case "-m":
			numberOfMoves = argsWithoutProg[i+1]
		case "-p":
			partition = argsWithoutProg[i+1]
		case "-t":
			timeLimit = argsWithoutProg[i+1]
		case "--help":
			fmt.Println("-l {Library IP}")
			fmt.Println("-m {Number of Moves}")
			fmt.Println("-t {Execution Time ex.5h}")
			fmt.Println("-p {Partition Name} (default: Auto Partition)")
			return
		}
	}

	if (numberOfMoves == "" && timeLimit == "") || libraryIP == ""{
		fmt.Println("Error not enough arguments / Use --help")
		return
	}

	var wg sync.WaitGroup
	var parsedNumMoves int
	var parsedTimeLimit time.Duration
	var err error

	if numberOfMoves != "" {
		parsedNumMoves, err = strconv.Atoi(numberOfMoves)
		if err != nil {
			fmt.Println("error parsing number of moves")
		}
	} else {
		parsedTimeLimit, err = time.ParseDuration(timeLimit)
		if err != nil {
			fmt.Println("error parsing time limit")
		}
	}


	successfulMoves := 0
	failedMoves := 0

	successfulMovesChan := make(chan int, 1)
	failedMovesChan := make(chan int, 1)

	successfulMovesChan <- successfulMoves
	failedMovesChan <- failedMoves

	currentTime := time.Now().Format("2006-01-02T15-04-05")
	logFileName := "moves_" + currentTime + ".log"
	file, err := os.OpenFile("./"+logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	log.SetOutput(file)
	rand.Seed(time.Now().UnixNano())
	log.Println("Library: " + libraryIP)
	log.Println("Partition: " + partition)

	if parsedNumMoves != 0 {
		log.Println("Move Count: " + numberOfMoves)
		for i := 0; i < parsedNumMoves; i++ {
			moveID, err := initMoveWorkflow(libraryIP, partition, (i+1))
			if err != nil {
				log.Println("#" + strconv.Itoa(i+1) + ": error: ", err)
				continue
			}
			go checkMoveStatus(&wg, libraryIP, moveID, successfulMovesChan, failedMovesChan)
			wg.Add(1)
		}
	} else {
		log.Println("Time Limit: " + parsedTimeLimit.String())
		timeStart := time.Now()
		count := 1
		for i := time.Since(timeStart); i < parsedTimeLimit; i = time.Since(timeStart) {
			moveID, err := initMoveWorkflow(libraryIP, partition, count)
			if err != nil {
				log.Println("#" + strconv.Itoa(count) + ": error: ", err)
				count++
				continue
			}
			go checkMoveStatus(&wg, libraryIP, moveID, successfulMovesChan, failedMovesChan)
			wg.Add(1)
			count++
		}
	}

	wg.Wait()

	log.Println("Moves Count: ", parsedNumMoves, " / Successful: ", <- successfulMovesChan, " / Failed: ", <- failedMovesChan)
}

func initMoveWorkflow(libraryIP, partition string, moveNum int) (string, error){
	inventory, err := getInventory(libraryIP, partition)
	if err != nil {
		return "", err
	}

	move, err := generateMove(partition, inventory)
	if err != nil {
		return "", err
	}

	moveID, err := sendMove(libraryIP, move)
	if err != nil {
		return "", err
	}

	log.Println("#" + strconv.Itoa(moveNum) + " - Move " + moveID + ": " + strconv.Itoa(move.Source) + " -> " + strconv.Itoa(move.Dest))
	return moveID, nil
}


