package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func getRandomAddr(addrPool []int) int {
	randMediaIndex := rand.Intn(len(addrPool))
	addr := addrPool[randMediaIndex]

	return addr
}

func getAddrByMode(mode int, populatedSlotAddr, emptySlotAddr, populatedDriveAddr, emptyDriveAddr []int) (int, int) {

	var source int
	var dest int

	switch mode {
	case 1:
		source = getRandomAddr(populatedSlotAddr)
		dest = getRandomAddr(emptySlotAddr)
	case 2:
		source = getRandomAddr(populatedSlotAddr)
		dest = getRandomAddr(emptyDriveAddr)
	case 3:
		source = getRandomAddr(populatedDriveAddr)
		dest = getRandomAddr(emptySlotAddr)
	}

	return source, dest
}

// Modes: Slot -> Slot : 1 / Slot -> Drive : 2 / Drive -> Slot : 3
func getMoveMode(populatedDriveAddr, emptyDriveAddr []int) int {
	moveMode := 1 + rand.Intn(3)

	// Check if Slot to Drive is possible
	if moveMode == 2 && len(emptyDriveAddr) == 0 {
		moveMode = 1
	}

	// Check if Drive to Slot is possible
	if moveMode == 3 && len(populatedDriveAddr) == 0 {
		moveMode = 1
	}

	// Default Slot to Slot if S -> D or D -> S not possible
	return moveMode
}

func generateMove(partition string, inventory []MediaContainer) (MoveRequest, error){
	populatedSlotAddr := []int{}
	emptySlotAddr := []int{}
	populatedDriveAddr := []int{}
	emptyDriveAddr := []int{}

	for _, v := range inventory {
		if v.MediaBarcode != nil {
			populatedSlotAddr = append(populatedSlotAddr, v.Address)
		} else {
			emptySlotAddr = append(emptySlotAddr, v.Address)
		}

		if v.Type == "Drive" {
			if v.MediaBarcode != nil {
				populatedDriveAddr = append(populatedDriveAddr, v.Address)
			} else {
				emptyDriveAddr = append(emptyDriveAddr, v.Address)
			}
		}
	}

	moveMode := getMoveMode(populatedDriveAddr, emptyDriveAddr)

	sourceAddr, destAddr := getAddrByMode(moveMode, populatedSlotAddr, emptySlotAddr, populatedDriveAddr, emptyDriveAddr)

	move := MoveRequest{
		Source:    sourceAddr,
		Dest:      destAddr,
		Partition: partition,
	}

	return move, nil
}

func sendMove(ip string, move MoveRequest) (string, error) {

	requestBody, err := json.Marshal(move)
	if err != nil {
		return "", err
	}

	resp, err := http.Post("http://" + ip + movesURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var moveResult Move
	json.Unmarshal(body, &moveResult)

	return strconv.FormatInt(moveResult.ID, 10), nil
}

func checkMoveStatus(wg *sync.WaitGroup, ip , moveID string, successCountChan, failCountChan chan int) {
	var moveStatus string

	for moveStatus != "Successful" && moveStatus != "Failed" {
		resp, err := http.Get("http://" + ip + movesURL + "/" + moveID)
		if err != nil {
			log.Println("error checking move status: ", err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("error sending move request: ", err)
		}

		var moveResult []Move
		err = json.Unmarshal(body, &moveResult)

		moveStatus = moveResult[0].Status
		time.Sleep(30 * time.Second)
	}

	if moveStatus == "Successful" {
		num := <- successCountChan
		num++
		successCountChan <- num
	} else {
		num := <- failCountChan
		num++
		failCountChan <- num
	}

	log.Println("Move " + moveID + " status: " + moveStatus)

	wg.Done()
}
