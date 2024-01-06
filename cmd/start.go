package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/GoToolSharing/htb-cli/config"
	"github.com/GoToolSharing/htb-cli/lib/utils"
	"github.com/GoToolSharing/htb-cli/lib/webhooks"
	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// coreStartCmd starts a specified machine and returns a status message and any error encountered.
func coreStartCmd(machineChoosen string, machineID string) (string, error) {
	var err error
	if machineID == "" {
		machineID, err = utils.SearchItemIDByName(machineChoosen, "Machine")
		if err != nil {
			return "", err
		}

	}
	config.GlobalConfig.Logger.Info(fmt.Sprintf("Machine ID: %s", machineID))

	machineType, err := utils.GetMachineType(machineID)
	if err != nil {
		return "", err
	}
	config.GlobalConfig.Logger.Info(fmt.Sprintf("Machine Type: %s", machineType))

	userSubscription, err := utils.GetUserSubscription()
	if err != nil {
		return "", err
	}
	config.GlobalConfig.Logger.Info(fmt.Sprintf("User subscription: %s", userSubscription))

	// isActive := utils.CheckVPN()
	// if !isActive {
	// 	isConfirmed := utils.AskConfirmation("No active VPN has been detected. Would you like to start it ?", batchParam)
	// 	if isConfirmed {
	// 		utils.StartVPN(config.BaseDirectory + "/lab_QU35T3190.ovpn")
	// 	}
	// }

	var url string
	var jsonData []byte

	switch {
	case machineType == "release":
		url = config.BaseHackTheBoxAPIURL + "/arena/start"
		jsonData = []byte("{}")
	case userSubscription == "vip" || userSubscription == "vip+":
		url = config.BaseHackTheBoxAPIURL + "/vm/spawn"
		jsonData, err = json.Marshal(map[string]string{"machine_id": machineID})
		if err != nil {
			return "", fmt.Errorf("failed to create JSON data: %w", err)
		}
	default:
		url = config.BaseHackTheBoxAPIURL + "/machine/play/" + machineID
		jsonData = []byte("{}")
	}

	resp, err := utils.HtbRequest(http.MethodPost, url, jsonData)
	if err != nil {
		return "", err
	}

	message, ok := utils.ParseJsonMessage(resp, "message").(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	if strings.Contains(message, "You must stop") {
		return message, nil
	}

	ip := "Undefined"
	switch {
	case userSubscription == "vip+" || machineType == "release":
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		setupSignalHandler(s)
		s.Suffix = " Waiting for the machine to start in order to fetch the IP address (this might take a while)."
		s.Start()
		defer s.Stop()
		timeout := time.After(10 * time.Minute)
	Loop:
		for {
			select {
			case <-timeout:
				fmt.Println("Timeout (10 min) ! Exiting")
				s.Stop()
				return "", nil
			default:
				ip, err = utils.GetActiveMachineIP()
				if err != nil {
					return "", err
				}
				if ip != "Undefined" {
					s.Stop()
					break Loop
				}
				time.Sleep(6 * time.Second)
			}
		}
	default:
		// Get IP address from active machine
		activeMachineData, err := utils.GetInformationsFromActiveMachine()
		if err != nil {
			return "", err
		}
		ip = activeMachineData["ip"].(string)
	}

	message = fmt.Sprintf("%s\nTarget: %s", message, ip)
	return message, nil
}

// startCmd defines the "start" command which initiates the starting of a specified machine.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a machine",
	Long:  `Starts a Hackthebox machine specified in argument`,
	Run: func(cmd *cobra.Command, args []string) {
		config.GlobalConfig.Logger.Info("Start command executed")
		machineChoosen, err := cmd.Flags().GetString("machine")
		if err != nil {
			config.GlobalConfig.Logger.Error("", zap.Error(err))
			os.Exit(1)
		}
		var machineID string
		if machineChoosen == "" {
			config.GlobalConfig.Logger.Info("Launching the machine in release arena")
			machineID, err = utils.SearchLastReleaseArenaMachine()
			if err != nil {
				config.GlobalConfig.Logger.Error("", zap.Error(err))
				os.Exit(1)
			}
			config.GlobalConfig.Logger.Debug(fmt.Sprintf("Machine ID : %s", machineID))

		}
		output, err := coreStartCmd(machineChoosen, machineID)
		if err != nil {
			config.GlobalConfig.Logger.Error("", zap.Error(err))
			os.Exit(1)
		}
		fmt.Println(output)
		err = webhooks.SendToDiscord("start", output)
		if err != nil {
			config.GlobalConfig.Logger.Error("", zap.Error(err))
			os.Exit(1)
		}
		config.GlobalConfig.Logger.Info("Exit start command correctly")
	},
}

// init adds the startCmd to rootCmd and sets flags for the "start" command.
func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringP("machine", "m", "", "Machine name")
}
