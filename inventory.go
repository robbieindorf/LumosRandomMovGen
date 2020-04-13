package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

func getInventory (ip, partitionName string) ([]MediaContainer, error){

	resp, err := http.Get("http://" + ip + inventoryURL + "/" + partitionName)
	if err != nil {
		return []MediaContainer{}, errors.New("unable to get inventory")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []MediaContainer{}, errors.New("error parsing inventory body")
	}

	var inventory []MediaContainer
	err = json.Unmarshal(body, &inventory)
	if err != nil {
		return []MediaContainer{}, errors.New("error unmarshalling inventory")
	}

	return inventory, nil
}
