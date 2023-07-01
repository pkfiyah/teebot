package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

const TEMP_PLAYERINFO_KEY = "LoadedUserInfo"

func LoadUserDataFromJson() error {
	file, err := os.Open("user.json")
	if err != nil {
		return err
	}

	defer file.Close()

	jsonBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = redClient.Set(ctx, TEMP_PLAYERINFO_KEY, jsonBytes, 0).Err()
	if err != nil {
		fmt.Printf("Err occurred saving user info to Redis: %v\n", err)
		return err
	}

	return nil
}

func GetPlayerInfo() (*PlayerInfo, error) {
	playerInfo := &PlayerInfo{}

	jsonBytes := redClient.Get(ctx, TEMP_PLAYERINFO_KEY)

	err := json.Unmarshal([]byte(jsonBytes.Val()), playerInfo)
	if err != nil {
		return nil, err
	}

	return playerInfo, nil
}

type PlayerInfo struct {
	User   UserInfo   `json:"userInfo"`
	Course CourseInfo `json:"courseInfo"`
}

type UserInfo struct {
	Username     string `json:"username"`
	Fullname     string `json:"fullname"`
	Password     string `json:"password"`
	PlayerID     string `json:"playerID"`
	LockerString string `json:"lockerString"`
}

type CourseInfo struct {
	CourseCode string `json:"courseCode"`
	Referrer   string `json:"referrer"`
}
