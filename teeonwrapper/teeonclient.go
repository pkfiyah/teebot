package teeonwrapper

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type TeeOnClient struct {
	client *http.Client
	jar    *cookiejar.Jar
	name   string
}

const teeOnSignInUrl string = "https://www.tee-on.com/PubGolf/servlet/com.teeon.teesheet.servlets.ajax.CheckSignInCloudAjax"
const teeOnTeeTimeUrl string = "https://www.tee-on.com/PubGolf/servlet/com.teeon.teesheet.servlets.golfersection.WebBookingBookTime"

const debug bool = true

func NewTeeOnClient() (*TeeOnClient, error) {
	toClient := TeeOnClient{}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	toClient.jar = jar
	toClient.client = &http.Client{
		Jar: jar,
	}

	toClient.name = "test"

	return &toClient, nil
}

func (toc *TeeOnClient) TeeOnSignIn() error {

	form := url.Values{}
	form.Set("Username", "Tylerfancy")
	form.Set("Password", "MansoN666")
	form.Set("SaveSignIn", "false")
	form.Set("CourseCode", "")

	req, err := http.NewRequest("POST", teeOnSignInUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	_, err = toc.client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func (toc *TeeOnClient) TeeOnSnipeTime(teeTime time.Time) error {
	// timestamping request
	nowTime := time.Now()
	unixT := nowTime.UnixMilli()

	// get tee time in way request wants it
	formatTeeTime := teeTime.Format("2006-01-02;15:04")

	parts := strings.Split(formatTeeTime, ";")
	if len(parts) != 2 {
		return errors.New("Request time could not be parsed properly into a tee time")
	}

	form := url.Values{}
	form.Set(fmt.Sprintf("%d-0", unixT), "Tyler Fancy")
	form.Set(fmt.Sprintf("%d-1", unixT), "Member")
	form.Set(fmt.Sprintf("%d-2", unixT), "Member")
	form.Set(fmt.Sprintf("%d-3", unixT), "Member")
	form.Set("BackTarget", "com.teeon.teesheet.servlets.golfersection.WebBookingPlayerEntry")
	form.Set("CaptureCreditBluff", "false")
	form.Set("CaptureCreditMoneris", "false")
	form.Set("Carts", "0")
	form.Set("CourseCode", "AVON")
	form.Set("Date", parts[0])
	form.Set("FromSpecials", "false")
	form.Set("Holes", "18")
	form.Set("LockerString", "Tyler Fancy (PUB281288)1|0")
	form.Set("Name0", "Tyler Fancy")
	form.Set("Name1", "Member")
	form.Set("Name2", "Member")
	form.Set("Name3", "Member")
	form.Set("NineCode", "F")
	form.Set("PlayerID0", "AVON3971")
	form.Set("PlayerID1", "")
	form.Set("PlayerID2", "")
	form.Set("PlayerID3", "")
	form.Set("Players", "4")
	form.Set("Referrer", "avonvalleygolf.com")
	form.Set("Ride0", "false")
	form.Set("Ride1", "false")
	form.Set("Ride2", "false")
	form.Set("Ride3", "false")
	form.Set("ShotgunID", "")
	form.Set("Time", parts[1])
	form.Set("UnlockTime", fmt.Sprintf("AVON|F|%s|%s|B|10:03|99", parts[0], parts[1]))

	req, err := http.NewRequest("POST", teeOnTeeTimeUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := toc.client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp.StatusCode != 200 {
			err = fmt.Errorf("Non 200 response occured: %d:%s", resp.StatusCode, resp.Body)
		}
		return fmt.Errorf("Error occured while booking tee time or received non-200 response: %s", err)
	}

	defer resp.Body.Close()
	_, err = scanResponseForSnipeMiss(resp.Body)

	return nil
}

func scanResponseForSnipeMiss(r io.Reader) (bool, error) {
	tooEarly, _ := regexp.Compile("You must wai[l-t]")
	notAvailable, _ := regexp.Compile("booking you requested is no longer available")

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if debug {
			fmt.Printf("[sRFSM]: %s\n", line)
		}
		if tooEarly.FindString(line) != "" {
			return true, nil
		}
		if notAvailable.FindString(line) != "" {
			return true, nil
		}
	}

	return false, nil
}
