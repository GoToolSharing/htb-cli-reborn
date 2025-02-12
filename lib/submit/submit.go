package submit

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/GoToolSharing/htb-cli/config"
	"github.com/GoToolSharing/htb-cli/lib/utils"
	"golang.org/x/term"
)

func SubmitFlag(url string, payload map[string]interface{}) (string, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to create JSON data: %w", err)
	}
	resp, err := utils.HtbRequest(http.MethodPost, url, jsonData)
	if err != nil {
		return "", err
	}

	message, ok := utils.ParseJsonMessage(resp, "message").(string)
	if !ok {
		return "", errors.New("unexpected response format")
	}
	return message, nil
}

// coreSubmitCmd handles the submission of flags for machines or challenges, returning a status message or error.
func CoreSubmitCmd(difficultyParam int, modeType string, modeValue string) (string, int, error) {
	var payload map[string]interface{}
	var difficultyString string
	var url string
	var challengeID string
	var mID int

	if modeType == "challenge" {
		config.GlobalConfig.Logger.Info("Challenge submit requested")
		if difficultyParam != 0 {
			if difficultyParam < 1 || difficultyParam > 10 {
				return "", 0, errors.New("difficulty must be set between 1 and 10")
			}
			difficultyString = strconv.Itoa(difficultyParam * 10)
		}
		challenges, err := utils.SearchChallengeByName(modeValue)
		if err != nil {
			return "", 0, err
		}
		config.GlobalConfig.Logger.Debug(fmt.Sprintf("Challenge found: %v", challenges))

		// TODO: get this int
		challengeID = strconv.Itoa(challenges.ID)

		url = config.BaseHackTheBoxAPIURL + "/challenge/own"
		payload = map[string]interface{}{
			"difficulty":   difficultyString,
			"challenge_id": challengeID,
		}
	} else if modeType == "machine" {
		config.GlobalConfig.Logger.Info("Machine submit requested")
		machineID, err := utils.SearchItemIDByName(modeValue, "Machine")
		if err != nil {
			return "", 0, err
		}
		machineData, err := utils.GetInformationsWithMachineId(machineID)
		if err != nil {
			return "", 0, err
		}
		if machineData["authUserInUserOwns"].(bool) && machineData["authUserInRootOwns"].(bool) {
			return "The machine has already been pwned", machineID, nil
		}

		machineType, err := utils.GetMachineType(machineID)
		if err != nil {
			return "", 0, err
		}
		config.GlobalConfig.Logger.Debug(fmt.Sprintf("Machine Type: %s", machineType))

		if machineType == "release" {
			url = config.BaseHackTheBoxAPIURL + "/arena/own"
		} else {
			url = config.BaseHackTheBoxAPIURL + "/machine/own"
		}
		payload = map[string]interface{}{
			"id": machineID,
		}
		mID = machineID
	} else if modeType == "fortress" {
		config.GlobalConfig.Logger.Info("Fortress submit requested")
		fortressID, err := utils.SearchFortressID(modeValue)
		if err != nil {
			return "", 0, err
		}
		config.GlobalConfig.Logger.Debug(fmt.Sprintf("Fortress ID : %d", fortressID))
		url = fmt.Sprintf("%s/fortress/%d/flag", config.BaseHackTheBoxAPIURL, fortressID)
		payload = map[string]interface{}{}
	} else if modeType == "prolab" {
		config.GlobalConfig.Logger.Info("Prolab submit requested")
		prolabID, err := utils.SearchProlabID(modeValue)
		if err != nil {
			return "", 0, err
		}
		config.GlobalConfig.Logger.Debug(fmt.Sprintf("Prolab ID : %d", prolabID))
		url = fmt.Sprintf("%s/prolab/%d/flag", config.BaseHackTheBoxAPIURL, prolabID)
		payload = map[string]interface{}{}
	} else if modeType == "active" {
		config.GlobalConfig.Logger.Info("Active machine submit requested")
		activeMachineData, err := utils.GetInformationsFromActiveMachine()
		if err != nil {
			return "", 0, err
		}
		if activeMachineData["authUserInUserOwns"].(bool) && activeMachineData["authUserInRootOwns"].(bool) {
			return "The machine has already been pwned", int(activeMachineData["id"].(float64)), nil
		}
		isConfirmed := utils.AskConfirmation("Would you like to submit a flag for the active machine ?")
		if !isConfirmed {
			return "", 0, nil
		}

		machineID, err := utils.GetActiveMachineID()
		if err != nil {
			return "", 0, err
		}
		config.GlobalConfig.Logger.Debug(fmt.Sprintf("Machine ID : %d", machineID))

		machineType, err := utils.GetMachineType(machineID)
		if err != nil {
			return "", 0, err
		}
		config.GlobalConfig.Logger.Debug(fmt.Sprintf("Machine Type: %s", machineType))

		if machineType == "release" {
			url = config.BaseHackTheBoxAPIURL + "/arena/own"
		} else {
			url = config.BaseHackTheBoxAPIURL + "/machine/own"
		}
		payload = map[string]interface{}{
			"id": machineID,
		}

		mID = machineID
	}

	fmt.Print("Flag : ")
	flagByte, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Error reading flag")
		return "", 0, fmt.Errorf("error reading flag")
	}
	flagOriginal := string(flagByte)
	flag := strings.ReplaceAll(flagOriginal, " ", "")

	config.GlobalConfig.Logger.Debug(fmt.Sprintf("Flag: %s", flag))

	payload["flag"] = flag

	message, err := SubmitFlag(url, payload)
	if err != nil {
		return "", 0, err
	}
	return message, mID, nil
}
