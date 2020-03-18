package MoveManager

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"lumosRandomMoves/models"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type MoveManager struct {
	 count  	int
	 ip			string
	 partition 	string
}


func (mm MoveManager) GenerateMove(partition string, inventory []models.MediaContainer) (models.MoveRequest, error){
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

	move := models.MoveRequest{
		Source:    sourceAddr,
		Dest:      destAddr,
		Partition: partition,
	}

	return move, nil
}

func (mm MoveManager) SendMove(ip string, move models.MoveRequest) (string, error) {

	requestBody, err := json.Marshal(move)
	if err != nil {
		return "", err
	}

	resp, err := http.Post("http://" + ip + models.movesURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var moveResult models.Move
	json.Unmarshal(body, &moveResult)

	return strconv.FormatInt(moveResult.ID, 10), nil
}

func (mm MoveManager) CheckMoveStatus(wg *sync.WaitGroup, moveID string, successCountChan, failCountChan chan int) {
	var moveStatus string

	for moveStatus != "Successful" && moveStatus != "Failed" {
		resp, err := http.Get("http://" + mm.ip + main.movesURL + "/" + moveID)
		if err != nil {
			log.Println("error checking move status: ", err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("error sending move request: ", err)
		}

		var moveResult []models.Move
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
