package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func getRandomAddr(addrPool []int) (int, error) {
	var addr int
	if len(addrPool) > 0 {
		randMediaIndex := rand.Intn(len(addrPool))
		addr = addrPool[randMediaIndex]
	} else {
		return 0, errors.New("error: no available addresses")
	}
	return addr, nil
}

func getAddrByMode(mode int, populatedSlotAddr, emptySlotAddr, populatedDriveAddr, emptyDriveAddr []int) (int, int, error) {

	var source int
	var dest int
	var err error

	switch mode {
	case 1:
		source, err = getRandomAddr(populatedSlotAddr)
		if err != nil {
			return 0, 0, errors.New(err.Error() + ": populated slot address")
		}
		dest, err = getRandomAddr(emptySlotAddr)
		if err != nil {
			return 0, 0, errors.New(err.Error() + ": empty slot address")
		}
	case 2:
		source, err = getRandomAddr(populatedSlotAddr)
		if err != nil {
			return 0, 0, errors.New(err.Error() + ": populated slot address")
		}
		dest, err = getRandomAddr(emptyDriveAddr)
		if err != nil {
			return 0, 0, errors.New(err.Error() + ": empty drive address")
		}
	case 3:
		source, err = getRandomAddr(populatedDriveAddr)
		if err != nil {
			return 0, 0, errors.New(err.Error() + ": populated drive address")
		}
		dest, err = getRandomAddr(emptySlotAddr)
		if err != nil {
			return 0, 0, errors.New(err.Error() + ": empty slot address")
		}
	}

	return source, dest, nil
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

	sourceAddr, destAddr, err := getAddrByMode(moveMode, populatedSlotAddr, emptySlotAddr, populatedDriveAddr, emptyDriveAddr)
	if err != nil {
		return MoveRequest{}, err
	}

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
	err = json.Unmarshal(body, &moveResult)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(moveResult.ID, 10), nil
}

func checkMoveStatus(wg *sync.WaitGroup, ip , moveID string, successCountChan, failCountChan chan int) {
	var moveStatus string
	checkCount := 0

	for moveStatus != "Successful" && moveStatus != "Failed" && checkCount < 5{
		resp, err := http.Get("http://" + ip + movesURL + "/" + moveID)
		if err != nil {
			log.Println("#" + moveID + " - Move" + moveID + " : error checking move status: ", err)
			checkCount++
			time.Sleep(30 * time.Second)
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("#" + moveID + " - Move" + moveID + " :error parsing move status body: ", err)
			checkCount++
			time.Sleep(30 * time.Second)
			continue
		}

		var moveResult []Move
		err = json.Unmarshal(body, &moveResult)
		if err != nil {
			log.Println("#" + moveID + " - Move" + moveID + " :error unmarshalling move status: ", err)
			checkCount++
			time.Sleep(30 * time.Second)
			continue
		}

		moveStatus = moveResult[0].Status
		checkCount++
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
