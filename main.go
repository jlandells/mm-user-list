package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
)

var Version = "development" // Default value - overwritten during bild process

var debugMode bool = false

// LogLevel is used to refer to the type of message that will be written using the logging code.
type LogLevel string

type mmConnection struct {
	mmURL    string
	mmPort   string
	mmScheme string
	mmToken  string
}

type User struct {
	UserID                string
	Username              string
	Email                 string
	FullName              string
	LastActivityOn        string
	DaysSinceLastActivity int
}

type MMUser struct {
	UserID                string
	Username              string
	Email                 string
	FirstName             string
	LastName              string
	Nickname              string
	IsBotAccount          bool
	UserCreatedAt         time.Time
	LastActivityAt        time.Time
	DaysSinceLastActivity int
	TeamName              string
}

const (
	debugLevel   LogLevel = "DEBUG"
	infoLevel    LogLevel = "INFO"
	warningLevel LogLevel = "WARNING"
	errorLevel   LogLevel = "ERROR"
)

const (
	defaultPort   = "8065"
	defaultScheme = "http"
	pageSize      = 60
	maxErrors     = 3
)

// Logging functions

// LogMessage logs a formatted message to stdout or stderr
func LogMessage(level LogLevel, message string) {
	if level == errorLevel {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(os.Stdout)
	}
	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("[%s] %s\n", level, message)
}

// DebugPrint allows us to add debug messages into our code, which are only printed if we're running in debug more.
// Note that the command line parameter '-debug' can be used to enable this at runtime.
func DebugPrint(message string) {
	if debugMode {
		LogMessage(debugLevel, message)
	}
}

// getEnvWithDefaults allows us to retrieve Environment variables, and to return either the current value or a supplied default
func getEnvWithDefault(key string, defaultValue interface{}) interface{} {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

// GetUsersNotInTeam returns a list of all Mattermost users who are without a team assignment
func GetUsersNotInTeam(mmClient *model.Client4, includeBots bool) ([]*MMUser, error) {

	DebugPrint("In GetUsersNotInTeam")

	ctx := context.Background()
	page := 0
	perPage := pageSize
	etag := ""

	var allUsers []*model.User

	for {
		users, response, err := mmClient.GetUsersWithoutTeam(ctx, page, perPage, etag)

		if err != nil {
			LogMessage(errorLevel, "Error returned from GetUsersWithoutTeam(): "+err.Error())
			return nil, err
		}
		if response.StatusCode != 200 {
			errMsg := fmt.Sprintf("Bad HTTP response returned from GetUsersWithoutTeam() (page %d)", page)
			LogMessage(errorLevel, errMsg)
			return nil, errors.New("failed to retrieve data from Mattermost")
		}

		if len(users) < perPage {
			allUsers = append(allUsers, users...)
			break
		}

		allUsers = append(allUsers, users...)

		page++
	}

	var userList []*MMUser

	for _, mmUser := range allUsers {
		if mmUser.IsBot && !includeBots {
			continue
		}
		userCreatedTime := time.Unix(0, mmUser.CreateAt*int64(time.Millisecond))
		lastActivityTime := time.Unix(0, mmUser.UpdateAt*int64(time.Millisecond))
		daysSinceLastActivity := int(time.Since(lastActivityTime).Hours() / 24)

		user := &MMUser{
			UserID:                mmUser.Id,
			Username:              mmUser.Username,
			Email:                 mmUser.Email,
			FirstName:             mmUser.FirstName,
			LastName:              mmUser.LastName,
			Nickname:              mmUser.Nickname,
			IsBotAccount:          mmUser.IsBot,
			UserCreatedAt:         userCreatedTime,
			LastActivityAt:        lastActivityTime,
			DaysSinceLastActivity: daysSinceLastActivity,
			TeamName:              "",
		}

		userList = append(userList, user)
	}

	return userList, nil
}

// GetUsersNotInTeam returns a list of all Mattermost users who are without a team assignment
func GetUsersInTeam(mmClient *model.Client4, team string, includeBots bool) ([]*MMUser, error) {

	DebugPrint("In GetUsersInTeam, for team: " + team)

	ctx := context.Background()
	page := 0
	perPage := pageSize
	etag := ""

	// First we need the team ID
	teams, response, err := mmClient.GetTeamByName(ctx, team, etag)

	if err != nil {
		LogMessage(errorLevel, "Error returned from GetTeamByName(): "+err.Error())
		return nil, err
	}
	if response.StatusCode != 200 {
		LogMessage(errorLevel, "Bad HTTP response returned from GetTeamByName()")
		return nil, errors.New("failed to retrieve data from Mattermost")
	}

	// There should only ever be one team retrieved
	teamID := teams.Id

	var allUsers []*model.User

	for {
		users, response, err := mmClient.GetUsersInTeam(ctx, teamID, page, perPage, etag)

		if err != nil {
			LogMessage(errorLevel, "Error returned from GetUsersInTeam(): "+err.Error())
			return nil, err
		}
		if response.StatusCode != 200 {
			errMsg := fmt.Sprintf("Bad HTTP response returned from GetUsersWithoutTeam() (page %d)", page)
			LogMessage(errorLevel, errMsg)
			return nil, errors.New("failed to retrieve data from Mattermost")
		}

		if len(users) < perPage {
			allUsers = append(allUsers, users...)
			break
		}

		allUsers = append(allUsers, users...)
		page++
	}

	var userList []*MMUser

	for _, mmUser := range allUsers {
		if mmUser.IsBot && !includeBots {
			continue
		}
		userCreatedTime := time.Unix(0, mmUser.CreateAt*int64(time.Millisecond))
		lastActivityTime := time.Unix(0, mmUser.UpdateAt*int64(time.Millisecond))
		daysSinceLastActivity := int(time.Since(lastActivityTime).Hours() / 24)

		user := &MMUser{
			UserID:                mmUser.Id,
			Username:              mmUser.Username,
			Email:                 mmUser.Email,
			FirstName:             mmUser.FirstName,
			LastName:              mmUser.LastName,
			Nickname:              mmUser.Nickname,
			IsBotAccount:          mmUser.IsBot,
			UserCreatedAt:         userCreatedTime,
			LastActivityAt:        lastActivityTime,
			DaysSinceLastActivity: daysSinceLastActivity,
			TeamName:              "",
		}

		userList = append(userList, user)
	}

	return userList, nil
}

func WriteUsersToCSV(users []*MMUser, filePath string) error {

	DebugPrint("Writing data to CSV file: " + filePath)

	file, err := os.Create(filePath)
	if err != nil {
		LogMessage(errorLevel, "Failed to create file: "+filePath+" - "+err.Error())
		return err
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the CSV header
	writer.Write([]string{
		"Username", "Email", "First Name", "Last Name", "Nickname", "Is Bot Account", "User Created Date",
		"Last Activity Date", "Days Since Last Activity", "Team Name",
	})

	// Iterate over the user data and write each record to the CSV file
	for _, user := range users {
		errorCount := 0
		record := []string{
			user.Username,
			user.Email,
			user.FirstName,
			user.LastName,
			user.Nickname,
			fmt.Sprintf("%v", user.IsBotAccount),          // Convert boolean to string.
			user.UserCreatedAt.Format("2006-01-02"),       // Format the time as a string.
			user.LastActivityAt.Format("2006-01-02"),      // Format the time as a string.
			fmt.Sprintf("%d", user.DaysSinceLastActivity), // Convert int to string.
			user.TeamName,
		}

		// Write the record to the CSV file
		if err := writer.Write(record); err != nil {
			LogMessage(warningLevel, "Failed to write record for user '"+user.Username+"' to CSV file")
			errorCount++
			if errorCount > maxErrors {
				LogMessage(errorLevel, "Too many errors writing to CSV file.  Aborting.")
				return err
			}
		}
	}

	return nil
}

func main() {

	// Parse Command Line
	DebugPrint("Parsing command line")

	var MattermostURL string
	var MattermostPort string
	var MattermostScheme string
	var MattermostToken string
	var MattermostTeam string
	var NotInTeam bool
	var IncludeBots bool
	var CSVFile string
	var DebugFlag bool
	var VersionFlag bool

	flag.StringVar(&MattermostURL, "url", "", "The URL of the Mattermost instance (without the HTTP scheme)")
	flag.StringVar(&MattermostPort, "port", "", "The TCP port used by Mattermost. [Default: "+defaultPort+"]")
	flag.StringVar(&MattermostScheme, "scheme", "", "The HTTP scheme to be used (http/https). [Default: "+defaultScheme+"]")
	flag.StringVar(&MattermostToken, "token", "", "The auth token used to connect to Mattermost")
	flag.StringVar(&MattermostTeam, "team", "", "The name of the Mattermost team")
	flag.BoolVar(&NotInTeam, "not-in-team", false, "Can be used in place of the 'team' parameter to only show users who are not allocated to a team.")
	flag.BoolVar(&IncludeBots, "include-bots", false, "Optional paramter to include bot accounts in the list")
	flag.StringVar(&CSVFile, "file", "", "*Required*  The name of the CSV file to which the output should be written")
	flag.BoolVar(&DebugFlag, "debug", false, "Enable debug output")
	flag.BoolVar(&VersionFlag, "version", false, "Show version information and exit")

	flag.Parse()

	if VersionFlag {
		fmt.Printf("\nmm-inactive-users - Version: %s\n\n", Version)
		os.Exit(0)
	}

	// If information not supplied on the command line, check whether it's available as an envrionment variable
	if MattermostURL == "" {
		MattermostURL = getEnvWithDefault("MM_URL", "").(string)
	}
	if MattermostPort == "" {
		MattermostPort = getEnvWithDefault("MM_PORT", defaultPort).(string)
	}
	if MattermostScheme == "" {
		MattermostScheme = getEnvWithDefault("MM_SCHEME", defaultScheme).(string)
	}
	if MattermostToken == "" {
		MattermostToken = getEnvWithDefault("MM_TOKEN", "").(string)
	}
	if !DebugFlag {
		DebugFlag = getEnvWithDefault("MM_DEBUG", debugMode).(bool)
	}

	DebugMessage := fmt.Sprintf("Parameters: \n  MattermostURL=%s\n  MattermostPort=%s\n  MattermostScheme=%s\n  MattermostToken=%s\n  Team=%s\n  CSV File=%s",
		MattermostURL,
		MattermostPort,
		MattermostScheme,
		MattermostToken,
		MattermostTeam,
		CSVFile)
	DebugPrint(DebugMessage)
	if NotInTeam {
		DebugPrint("'Not In Team' flag is set")
	}
	if IncludeBots {
		DebugPrint("'Include Bots' flag is set")
	}

	// Validate required parameters
	DebugPrint("Validating parameters")
	var cliErrors bool = false
	if MattermostURL == "" {
		LogMessage(errorLevel, "The Mattermost URL must be supplied either on the command line of vie the MM_URL environment variable")
		cliErrors = true
	}
	if MattermostScheme == "" {
		LogMessage(errorLevel, "The Mattermost HTTP scheme must be supplied either on the command line of vie the MM_SCHEME environment variable")
		cliErrors = true
	}
	if MattermostToken == "" {
		LogMessage(errorLevel, "The Mattermost auth token must be supplied either on the command line of vie the MM_TOKEN environment variable")
		cliErrors = true
	}
	// if MattermostTeam == "" {
	// 	LogMessage(errorLevel, "A Mattermost team name is required to use this utility.")
	// 	cliErrors = true
	// }
	if CSVFile == "" {
		LogMessage(errorLevel, "A CSV output file must be specified")
		cliErrors = true
	}
	if MattermostTeam != "" && NotInTeam {
		LogMessage(errorLevel, "Only one of 'team' or 'not-in-teams' can be specified")
		cliErrors = true
	}
	if cliErrors {
		flag.Usage()
		os.Exit(1)
	}

	debugMode = DebugFlag

	mattermostConenction := mmConnection{
		mmURL:    MattermostURL,
		mmPort:   MattermostPort,
		mmScheme: MattermostScheme,
		mmToken:  MattermostToken,
	}

	mmTarget := fmt.Sprintf("%s://%s:%s", mattermostConenction.mmScheme, mattermostConenction.mmURL, mattermostConenction.mmPort)

	DebugPrint("Full target for Mattermost: " + mmTarget)
	mmClient := model.NewAPIv4Client(mmTarget)
	mmClient.SetToken(mattermostConenction.mmToken)
	DebugPrint("Connected to Mattermost")

	LogMessage(infoLevel, "Processing started - Version: "+Version)

	var users []*MMUser
	var err error

	if NotInTeam {
		users, err = GetUsersNotInTeam(mmClient, IncludeBots)
	} else {
		if MattermostTeam == "" {
			LogMessage(errorLevel, "Mattermost team is required!")
			flag.Usage()
			os.Exit(3)
		}
		users, err = GetUsersInTeam(mmClient, MattermostTeam, IncludeBots)
	}
	if err != nil {
		LogMessage(errorLevel, "Processing failed.  Error: "+err.Error())
		os.Exit(2)
	}

	if len(users) > 0 {
		err := WriteUsersToCSV(users, CSVFile)
		if err != nil {
			LogMessage(errorLevel, "Failed to create CSV file: "+err.Error())
			os.Exit(4)
		}
	} else {
		LogMessage(warningLevel, "No users found to write to CSV!")
	}

}
