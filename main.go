package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	gomail "gopkg.in/gomail.v2"
)

/*
"fridges" (the esp8266s) will do the following:
    -   request at /api/status every 5 minutes
        -   with name, temp, humidity, and isOpen
    -   request at /api/open after each open
        -   with name, seconds, temp, and humidity

env vars:
	-	MAIL_EMAIL
	-	MAIL_PASSWORD
	-	PORT
	-	ADMIN_EMAIL
*/

const (
	FridgeListFile      = "./fridgeList.json"
	FridgeEmailsFile    = "./fridgeEmails.json"
	TimersFile          = "./timers.json"
	OpenLogsPath        = "./webroot/openLogs/"
	StatusLogsPath      = "./webroot/statusLogs/"
	saveTimerMinutes    = 5
	TimeBeforeEmergency = 1800
	MailServer          = "smtp.gmail.com"
	MailPort            = 587
)

// constant as well
var statusCSVOrder []string = []string{"Date", "Time", "Temp", "Humidity"}
var openCSVOrder []string = []string{"Seconds", "Date", "Time", "Temp", "Humidity"}

// public facing json, only fridge names
var fridgeList []string

// private json, names linked to emails for notifications
var emailsList []fridgeEmails

// timers, to track how long it has been since last request
var timers []timer

// runtime variable, doesn't get saved to json
var runningTimers []runningTimer

// timer json is periodically saved, this inhibits in case another function is saving to file
var inhibitSave bool = false

type timer struct {
	// name should be normalized
	Name                   string `json:"name"`
	Timer                  int64  `json:"timer"`
	IsDown                 bool   `json:"isdown"`
	HasDoneEmergencyAction bool   `json:"hasdoneemergaction"`
}

type runningTimer struct {
	name      string
	index     int
	isRunning bool
}

type fridgeEmails struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func getFridgeEmail(name string) (bool, int, string) {
	for index, pair := range emailsList {
		if nstr(pair.Name) == nstr(name) {
			return true, index, pair.Email
		}
	}
	return false, 0, ""
}

func emergencyHandler(name string) {
	isInList, _, email := getFridgeEmail(name)
	fmt.Println("Emergency handler function!!")
	msg := gomail.NewMessage()
	msg.SetHeader("From", os.Getenv("MAIL_EMAIL"))
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Lost contact with a community fridge")
	msg.SetBody("text/html", `<hr><h2>Lost contact with a community fridge.</h2><br><p>Community fridge, named "`+name+`", has not sent a status update to The Food Grid's servers for 30 minutes.`)
	if !isInList {
		fmt.Println("Email for " + name + " is not configured")
	} else {
		fmt.Println("Contacting owner of community fridge " + name)
		if os.Getenv("MAIL_EMAIL") == "" || os.Getenv("MAIL_PASSWORD") == "" {
			fmt.Println("MAIL_EMAIL and/or MAIL_PASSWORD not set. Not sending email")
		} else {
			msg.SetHeader("From", os.Getenv("MAIL_EMAIL"))
			msg.SetHeader("To", email)
			msg.SetHeader("Subject", "Lost contact with your community fridge")
			msg.SetBody("text/html", `<img src="https://foodgridia.org/wp-content/uploads/2020/11/FGI_Logo_350x280-White.png" width="350" height="280" style="display:block;border:0;color:black;font-size:25px;font-weight:bold;font-family:sans-serif;" /><hr><h2>We have lost contact with your community fridge.</h2><br><p>Your community fridge, named "`+name+`", has not sent a status update to The Food Grid's servers for 30 minutes. We recommend that you check on your fridge.</p>`)
			msg.SetHeader("To", email)
			n := gomail.NewDialer(MailServer, MailPort, os.Getenv("MAIL_EMAIL"), os.Getenv("MAIL_PASSWORD"))
			err := n.DialAndSend(msg)
			if err != nil {
				fmt.Println("Error sending email, " + err.Error())
			} else {
				fmt.Println("Successfully sent email")
			}
		}
	}
	if os.Getenv("ADMIN_EMAIL") != "" && os.Getenv("MAIL_EMAIL") != "" && os.Getenv("MAIL_PASSWORD") != "" {
		fmt.Println("Sending admin an email")
		msg.SetHeader("To", os.Getenv("ADMIN_EMAIL"))
		n := gomail.NewDialer(MailServer, MailPort, os.Getenv("MAIL_EMAIL"), os.Getenv("MAIL_PASSWORD"))
		err := n.DialAndSend(msg)
		if err != nil {
			fmt.Println("Error sending admin email, " + err.Error())
		} else {
			fmt.Println("Successfully sent admin email")
		}
	}
}

func loadNamesJSON() {
	fileBytes, err := os.ReadFile(FridgeListFile)
	if err == nil {
		err := json.Unmarshal(fileBytes, &fridgeList)
		if err != nil {
			fmt.Println("Error unmarshaling names JSON")
			fmt.Println(err)
			return
		}
		fmt.Println("Names JSON loaded:")
		fmt.Println(fridgeList)
		return
	}
	os.Create(FridgeListFile)
}

func loadEmailsJSON() {
	fileBytes, err := os.ReadFile(FridgeEmailsFile)
	if err == nil {
		err := json.Unmarshal(fileBytes, &emailsList)
		if err != nil {
			fmt.Println("Error unmarshaling emails JSON (expected if none)")
			fmt.Println(err)
			return
		}
		fmt.Println("Emails JSON loaded:")
		fmt.Println(emailsList)
		return
	}
	os.Create(FridgeEmailsFile)
}

func loadTimersJSON() {
	fileBytes, err := os.ReadFile(TimersFile)
	if err == nil {
		err := json.Unmarshal(fileBytes, &timers)
		if err != nil {
			fmt.Println("Error unmarshaling timers JSON")
			fmt.Println(err)
			return
		}
		for index, timer := range timers {
			if !timer.IsDown {
				startTimer(index)
			} else {
				fmt.Println(timer.Name + " is in timers file, but it is reported to be down so we are not starting a timer for it")
			}
		}
		fmt.Println("Timers JSON loaded:")
		fmt.Println(timers)
		return
	}
	os.Create(TimersFile)
}

func addEmail(name, email string) {
	isInList, index, oldEmail := getFridgeEmail(name)
	if isInList {
		if oldEmail == email {
			fmt.Println("Add email request made, but email is already in JSON")
			return
		} else {
			emailsList[index].Email = email
			fileBytes, _ := json.Marshal(emailsList)
			fmt.Println("Changing email in list")
			os.WriteFile(FridgeEmailsFile, fileBytes, 0644)
		}
	} else {
		emailsList = append(emailsList, fridgeEmails{nstr(name), email})
		fileBytes, _ := json.Marshal(emailsList)
		fmt.Println("Adding an email to list")
		os.WriteFile(FridgeEmailsFile, fileBytes, 0644)
	}
}

// normalize a string for use in comparisons and file output
// ex: input=" Community Youth Concepts", output="Community_Youth_Concepts"
func nstr(in string) string {
	return strings.Replace(strings.TrimSpace(in), " ", "_", -1)
}

func addFridge(in string) {
	matched := false
	for _, name := range fridgeList {
		if nstr(name) == nstr(in) {
			matched = true
			break
		}
	}
	if matched {
		return
	} else {
		fmt.Println("New fridge, adding " + in + " to fridge list")
		fridgeList = append(fridgeList, in)
		writeBytes, _ := json.Marshal(fridgeList)
		os.WriteFile(FridgeListFile, writeBytes, 0644)
	}
}

func removeFridge(in string) error {
	matched := false
	for index, name := range fridgeList {
		if nstr(name) == nstr(in) {
			matched = true
			fmt.Println("Deleting " + in + " from fridge list")
			fridgeList = append(fridgeList[:index], fridgeList[index+1:]...)
			break
		}
	}
	if !matched {
		return fmt.Errorf(in + " not in fridge list")
	} else {
		writeBytes, _ := json.Marshal(fridgeList)
		fmt.Println("Removing " + in + " from fridge list")
		os.WriteFile(FridgeListFile, writeBytes, 0644)
	}
	return nil
}

func makeStatusEntry(name, temp, humidity, isOpen string) {
	now := time.Now()
	date := fmt.Sprintf("%02d-%02d-%d", now.Month(), now.Day(), now.Year())
	time := now.Format("15:04")
	entry := []string{date, time, temp, humidity}
	filename := StatusLogsPath + nstr(name) + ".csv"
	fileBytes, err := os.ReadFile(filename)
	if err == nil {
		file := bytes.NewBuffer(fileBytes)
		r := csv.NewWriter(file)
		r.Write(entry)
		r.Flush()
		os.WriteFile(filename, file.Bytes(), 0644)
	} else {
		// create file
		fmt.Println("Creating new status file for fridge " + name)
		file, _ := os.Create(filename)
		r := csv.NewWriter(file)
		r.Write(statusCSVOrder)
		r.Write(entry)
		r.Flush()
	}
}

func makeOpenEntry(name, seconds, temp, humidity string) {
	now := time.Now()
	date := fmt.Sprintf("%02d-%02d-%d", now.Month(), now.Day(), now.Year())
	time := now.Format("15:04")
	entry := []string{seconds, date, time, temp, humidity}
	filename := OpenLogsPath + nstr(name) + ".csv"
	fileBytes, err := os.ReadFile(filename)
	if err == nil {
		file := bytes.NewBuffer(fileBytes)
		r := csv.NewWriter(file)
		r.Write(entry)
		r.Flush()
		os.WriteFile(filename, file.Bytes(), 0644)
	} else {
		// create file
		fmt.Println("Creating new opens file for fridge " + name)
		file, _ := os.Create(filename)
		r := csv.NewWriter(file)
		r.Write(openCSVOrder)
		r.Write(entry)
		r.Flush()
	}
}

func getTimer(in string) (timer, int, bool) {
	matched := false
	for index, timer := range timers {
		if nstr(in) == nstr(timer.Name) {
			matched = true
			return timer, index, matched
		}
	}
	return timer{}, 0, false
}

func getRunningTimer(in int) (runningTimer, int, bool) {
	matched := false
	for index, runningTimer := range runningTimers {
		if in == runningTimer.index {
			matched = true
			return runningTimer, index, matched
		}
	}
	return runningTimer{}, 0, false
}

func saveTimer() {
	go func() {
		for {
			time.Sleep(time.Minute * saveTimerMinutes)
			//time.Sleep(time.Second * 10)
			if !inhibitSave {
				timerBytes, err := json.Marshal(timers)
				if err == nil {
					os.WriteFile(TimersFile, timerBytes, 0644)
				} else {
					fmt.Println("Error saving timer file")
					fmt.Println(err)
				}
			}
		}
	}()
}

func startTimer(timerIndex int) {
	_, index, isInList := getRunningTimer(timerIndex)
	if !isInList {
		fmt.Println(timers[timerIndex].Name + " is not in running timers list, appending")
		newRunningTimer := runningTimer{timers[timerIndex].Name, timerIndex, false}
		runningTimers = append(runningTimers, newRunningTimer)
		index = len(runningTimers) - 1
	}
	if !runningTimers[index].isRunning {
		fmt.Println("Starting new timer")
		runningTimers[index].isRunning = true
		timers[timerIndex].HasDoneEmergencyAction = false
		timers[timerIndex].Timer = 0
		go func() {
			for {
				timers[timerIndex].Timer = timers[timerIndex].Timer + 1
				time.Sleep(time.Second)
				if timers[timerIndex].Timer >= TimeBeforeEmergency {
					fmt.Println(timers[timerIndex].Name + " has not made a request for 30 minutes. Initiating emergency function")
					emergencyHandler(timers[timerIndex].Name)
					timers[timerIndex].IsDown = true
					runningTimers[index].isRunning = false
					timers[timerIndex].HasDoneEmergencyAction = true
					inhibitSave = true
					timerBytes, err := json.Marshal(timers)
					if err == nil {
						os.WriteFile(TimersFile, timerBytes, 0644)
					} else {
						fmt.Println("Error saving timer file")
						fmt.Println(err)
					}
					inhibitSave = false
					return
				}
			}
		}()
	} else {
		timers[timerIndex].Timer = 0
	}
}

func timerHandler(name string) {
	_, index, isTimer := getTimer(name)
	if !isTimer {
		inhibitSave = true
		fmt.Println(name + " is not in timers list, appending")
		newTimer := timer{nstr(name), 0, false, false}
		timers = append(timers, newTimer)
		index = len(timers) - 1
		timerBytes, err := json.Marshal(timers)
		if err == nil {
			os.WriteFile(TimersFile, timerBytes, 0644)
		} else {
			fmt.Println("Error saving timer file")
			fmt.Println(err)
		}
		inhibitSave = false
	}
	startTimer(index)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	default:
		http.Error(w, "not found", http.StatusNotFound)
		return
	case r.URL.Path == "/api/open":
		name := r.FormValue("name")
		seconds := r.FormValue("seconds")
		temp := r.FormValue("temp")
		humidity := r.FormValue("humidity")
		addFridge(name)
		makeOpenEntry(name, seconds, temp, humidity)
		fmt.Fprintf(w, "ok")
		return
	case r.URL.Path == "/api/status":
		name := r.FormValue("name")
		temp := r.FormValue("temp")
		humidity := r.FormValue("humidity")
		isOpen := r.FormValue("isOpen")
		addFridge(name)
		makeStatusEntry(name, temp, humidity, isOpen)
		timerHandler(name)
		fmt.Fprintf(w, "ok")
		return
	case r.URL.Path == "/api/delete":
		name := r.FormValue("name")
		if name == "" {
			fmt.Fprintf(w, "error: must provide a fridge name")
			return
		}
		err := removeFridge(name)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprintf(w, "ok")
		return
	case r.URL.Path == "/api/get_json":
		fileBytes, err := os.ReadFile(FridgeListFile)
		if err != nil {
			fmt.Fprintf(w, "error: "+err.Error())
			return
		}
		w.Write(fileBytes)
		return
	case r.URL.Path == "/api/save_email":
		name := r.FormValue("name")
		email := r.FormValue("email")
		if name == "" || email == "" {
			fmt.Fprintf(w, "error: you must provide a name and email")
			return
		}
		addEmail(name, email)
		fmt.Fprintf(w, "ok")
		return
	case r.URL.Path == "/api/get_timers":
		returnBytes, err := json.Marshal(timers)
		if err != nil {
			fmt.Fprintf(w, "error: "+err.Error())
			return
		}
		w.Write(returnBytes)
		return
	case r.URL.Path == "/api/print_running_timers":
		fmt.Fprintln(w, runningTimers)
		fmt.Println(runningTimers)
		return
	}

}

func main() {
	loadNamesJSON()
	loadEmailsJSON()
	loadTimersJSON()
	saveTimer()
	http.HandleFunc("/api/", apiHandler)
	http.Handle("/", http.FileServer(http.Dir("./webroot")))
	var port string
	if os.Getenv("PORT") == "" {
		port = "83"
	} else {
		port = os.Getenv("PORT")
	}

	fmt.Println("Starting server at port " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
