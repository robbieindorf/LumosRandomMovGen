package MoveManager

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"lumosRandomMoves"
	"lumosRandomMoves/models"
	"net/http"
)

func getInventory (ip, partitionName string) ([]models.MediaContainer, error){

	resp, err := http.Get("http://" + ip + main.inventoryURL + "/" + partitionName)
	if err != nil {
		return []models.MediaContainer{}, errors.New("unable to get inventory")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []models.MediaContainer{}, errors.New("error parsing inventory")
	}

	var inventory []models.MediaContainer
	err = json.Unmarshal(body, &inventory)
	if err != nil {
		return []models.MediaContainer{}, errors.New("error parsing request body")
	}

	return inventory, nil
}
