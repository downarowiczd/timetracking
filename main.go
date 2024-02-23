package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"downardo.at/timetracking/internal/domain"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

const VERSION = "0.0.1"

var TrackingRepositroy *domain.SQLiteRepository

func initConfig() {
	// Setting up some configurations
	viper.Set("databaseFile", "recordings.db")
	viper.Set("databaseDriver", "sqlite")
	// Creating and updating a configuration file
	viper.SetConfigName("app_config") // name of config file (without extension)
	viper.SetConfigType("yaml")       // specifying the config type
	viper.AddConfigPath(".")          // path to look for the config file in

	err := viper.SafeWriteConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileAlreadyExistsError); ok {
			err = viper.WriteConfig()
			if err != nil {
				log.Fatalf("Error while updating config file %s", err)
			}
		} else {
			log.Fatalf("Error while creating config file %s", err)
		}
	}

	log.Print("Configuration file created/updated successfully!")
}

func initDatabase() *domain.SQLiteRepository {
	db, err := sql.Open(viper.GetString("databaseDriver"), viper.GetString("databaseFile"))
	if err != nil {
		log.Fatal(err)
	}

	// Migrate the database
	trackingRepositroy := domain.NewSQLiteRepository(db)

	if err := trackingRepositroy.Migrate(); err != nil {
		log.Fatal(err)
	}
	return trackingRepositroy
}

func Info(a ...interface{}) (int, error) {
	return color.New(color.FgWhite, color.Bold).Println(a...)
}

func Notice(a ...interface{}) (int, error) {
	return color.New(color.Bold, color.FgGreen).Println(a...)
}

func InputPrint() {
	color.New(color.Bold).Print("Enter command: -> ")
	//fmt.Print("Enter command: -> ")
}

func clearTerminal() {
	fmt.Print("\033[H\033[2J")
}

func pressEnterToContinue() {
	Info("Press enter to continue ...")

	// Wait for the user to press enter and then clear the terminal
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func printTopBar() {
	Info("=================================")
	tn := time.Now()
	year, week := tn.ISOWeek()
	Info("Time Tracking Week: ", week, " Year: ", year)
	//Get today's date

}

func main() {
	clearTerminal()
	// Create a custom print function for convenience
	//red := color.New(color.Bold, color.FgWhite, color.BgHiRed).PrintlnFunc()
	// Mix up multiple attributes

	Info("Time Tracking version: ", VERSION)
	log.Print("Initializing time tracking ...")
	initConfig()
	log.Print("Config initialized successfully")
	log.Print("Initializing database ...")
	TrackingRepositroy := initDatabase()
	log.Print("Database initialized successfully")

	Info("---------------------------------")
	Notice("Time Tracking is ready to use!")
	Info("---------------------------------")

	pressEnterToContinue()
	clearTerminal()

	for {
		clearTerminal()
		printTopBar()
		InputPrint()
		text, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		// convert CRLF to LF
		// for Windows
		text = strings.Replace(text, "\r\n", "", -1)
		// for Linux
		text = strings.Replace(text, "\n", "", -1)

		fmt.Println(text)
		if text == "help" {
			Info("Available commands:")
			Info(" week: Show the current week's recordings")
			Info("  - stop: Stop a recording")
			Info("  - list: List all recordings")
			Info("  - exit: Exit the application")
			TrackingRepositroy.CreateRecording(domain.Recording{ProjectTag: "test", Name: "test", Billable: true, Status: 0})
			pressEnterToContinue()
		} else if text == "week" {
			clearTerminal()
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"#", "First Name", "Last Name", "Salary"})
			t.AppendRows([]table.Row{
				{1, "Arya", "Stark", 3000},
				{20, "Jon", "Snow", 2000, "You know nothing, Jon Snow!"},
			})
			t.AppendSeparator()
			t.AppendRow([]interface{}{300, "Tyrion", "Lannister", 5000})
			t.AppendFooter(table.Row{"", "", "Total", 10000})
			t.SetStyle(table.StyleColoredBright)
			t.Render()

			pressEnterToContinue()
		} else if text == "week matrix" {
			clearTerminal()
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"#", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"})
			t.AppendRows([]table.Row{
				{"DAG", 4.25, 2.00, 1.00, 3.00, 5.00, 0.00, 0.00},
				{"INT", 10.00, 4.00, 7.00, 6.00, 8.00, 0.00, 0.00},
			})
			t.AppendSeparator()
			t.AppendRow([]interface{}{"#", 16.25, 6.00, 8.00, 9.00, 13.00, 0.00, 0.00})
			t.AppendFooter(table.Row{"Total week", 52.25})
			t.SetStyle(table.StyleColoredBright)
			t.Render()

			pressEnterToContinue()
		} else if text == "exit" {
			break
		}
	}

}
