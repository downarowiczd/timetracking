package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"downardo.at/timetracking/internal/domain"
	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

const VERSION = "0.0.1"

var TrackingRepositroy *domain.SQLiteRepository

func initConfig() {
	// Setting up some configurations

	// Creating and updating a configuration file
	viper.SetConfigName("app_config") // name of config file (without extension)
	viper.SetConfigType("yaml")       // specifying the config type
	viper.AddConfigPath(".")          // path to look for the config file in

	viper.ReadInConfig() // Find and read the config file

	// set default values if they are unset
	if viper.GetString("databaseDriver") == "" {
		viper.SetDefault("databaseDriver", "sqlite")
		viper.SetDefault("databaseFile", "timetracking.db")
		//if there is no config file, create one
		if err := viper.WriteConfigAs("app_config.yaml"); err != nil {
			log.Fatal(err)
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

func printProjectList(repo *domain.SQLiteRepository, onlyActive bool) {
	clearTerminal()

	projects, err := repo.AllProjects()
	if onlyActive {
		projects, err = repo.AllActiveProjects()
	}
	if err != nil {
		log.Fatal(err)
	}
	if onlyActive {
		Notice("Project List - Active Projects")
	} else {
		Notice("Project List - All Projects")
	}
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Tag", "Name", "Type", "Status"})

	for _, project := range projects {
		t.AppendRow([]interface{}{project.Tag, project.Name, project.Type, project.StatusString()})
	}
	t.SetStyle(table.StyleDouble)
	t.Render()
	Info("Available commands: [new, edit (tag), delete (tag), all, active, exit] [tag]")
	InputPrint()
}

func projectMenu(repo *domain.SQLiteRepository) {
	clearTerminal()
	onlyActive := true
	for {
		printProjectList(repo, onlyActive)
		text, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		// convert CRLF to LF
		// for Windows
		text = strings.Replace(text, "\r\n", "", -1)
		// for Linux
		text = strings.Replace(text, "\n", "", -1)
		if text == "new" {
			clearTerminal()
			addProjectForm(repo)
			printProjectList(repo, onlyActive)
		} else if strings.HasPrefix(text, "edit") {
			//get the tag
			args := strings.Split(text, " ")
			if len(args) < 2 {
				Info("Please enter a tag")
				pressEnterToContinue()
				printProjectList(repo, onlyActive)
			} else {
				tag := args[1]
				if tag == "" {
					Info("Please enter a tag")
					pressEnterToContinue()
					printProjectList(repo, onlyActive)
				} else {
					if _, err := repo.GetProjectByTag(tag); err != nil {
						Info("Project not found")
						pressEnterToContinue()
						printProjectList(repo, onlyActive)
					} else {
						editProjectForm(repo, tag)
						printProjectList(repo, onlyActive)
					}
				}
			}
		} else if strings.HasPrefix(text, "delete") {
			//get the tag
			args := strings.Split(text, " ")
			if len(args) < 2 {
				Info("Please enter a tag")
				pressEnterToContinue()
			} else {
				tag := args[1]
				if tag == "" {
					Info("Please enter a tag")
					pressEnterToContinue()
					printProjectList(repo, onlyActive)
				} else {
					if _, err := repo.GetProjectByTag(tag); err != nil {
						Info("Project not found")
						pressEnterToContinue()
						printProjectList(repo, onlyActive)
					} else {
						err := repo.DeleteProject(tag)
						if err != nil {
							log.Fatal(err)
						} else {
							clearTerminal()
							Info("Project deleted successfully!")
							pressEnterToContinue()
						}
					}
				}
			}
		} else if strings.HasPrefix(text, "exit") {
			break
		} else if text == "all" {
			onlyActive = false
		} else if text == "active" {
			onlyActive = true
		} else {
			Info("Invalid command")
			pressEnterToContinue()
		}
	}
}

func addProjectForm(repo *domain.SQLiteRepository) {
	var (
		tag         string
		name        string
		projectType string
		status      string
		confirm     bool
	)
	form := huh.NewForm(
		// Gather some final details about the order.
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				CharLimit(25).
				Value(&name).
				// Validating fields is easy. The form will mark erroneous fields
				// and display error messages accordingly.
				Validate(func(str string) error {
					if str == "" {
						return errors.New("please enter a name.")
					}
					return nil
				}),

			huh.NewInput().
				Title("Project tag").
				CharLimit(10).
				Value(&tag).
				// Validating fields is easy. The form will mark erroneous fields
				// and display error messages accordingly.
				Validate(func(str string) error {
					if str == "" {
						return errors.New("please enter a tag.")
					} else if len(str) > 10 {
						return errors.New("tag is too long")
					} else {
						_, err := repo.GetProjectByTag(str)
						if err == nil {
							return errors.New("tag already exists")
						}
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Project type").
				Options(
					huh.NewOption("Internal", "internal"),
					huh.NewOption("Customer", "customer"),
					huh.NewOption("Development", "development"),
					huh.NewOption("Open Source", "open source"),
					huh.NewOption("Other", "other"),
				).
				Value(&projectType),

			huh.NewSelect[string]().
				Title("Status").
				Options(
					huh.NewOption("Active", "0"),
					huh.NewOption("Inactive", "1"),
				).
				Value(&status),
			huh.NewConfirm().
				Title("Create new project?").
				Affirmative("Yes!").
				Negative("No.").
				Value(&confirm),
		),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	} else {
		if confirm == true {
			Info("Creating project ...")
			project := domain.Project{
				Tag:    tag,
				Name:   name,
				Type:   projectType,
				Status: 0,
			}
			if status == "1" {
				project.Status = 1
			}
			_, err := repo.CreateProject(project)
			if err != nil {
				log.Fatal(err)
			}
			Info("Project created successfully!")
		} else {
			Info("Project creation canceled")
			return
		}
	}
}

func editProjectForm(repo *domain.SQLiteRepository, tag string) {
	var (
		name        string
		projectType string
		status      string
		confirm     bool
	)

	project, err := repo.GetProjectByTag(tag)
	if err != nil {
		log.Fatal(err)
	}

	projectType = project.Type
	status = fmt.Sprintf("%d", project.Status)
	name = project.Name

	form := huh.NewForm(
		// Gather some final details about the order.
		huh.NewGroup(
			huh.NewNote().
				Title("Edit project '"+tag+"'"),
			huh.NewInput().
				Title("Project name").
				CharLimit(25).
				Value(&name).
				// Validating fields is easy. The form will mark erroneous fields
				// and display error messages accordingly.
				Validate(func(str string) error {
					if str == "" {
						return errors.New("please enter a name.")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Project type").
				Options(
					huh.NewOption("Internal", "internal"),
					huh.NewOption("Customer", "customer"),
					huh.NewOption("Development", "development"),
					huh.NewOption("Open Source", "open source"),
					huh.NewOption("Other", "other"),
				).
				Value(&projectType),

			huh.NewSelect[string]().
				Title("Status").
				Options(
					huh.NewOption("Active", "0"),
					huh.NewOption("Inactive", "1"),
				).
				Value(&status),
			huh.NewConfirm().
				Title("Create new project?").
				Affirmative("Yes!").
				Negative("No.").
				Value(&confirm),
		),
	)

	errForm := form.Run()
	if errForm != nil {
		log.Fatal(errForm)
	} else {
		if confirm == true {
			Info("Editing project " + tag + " ...")
			project := domain.Project{
				Tag:    tag,
				Name:   name,
				Type:   projectType,
				Status: 0,
			}
			if status == "1" {
				project.Status = 1
			}
			_, err := repo.UpdateProject(tag, project)
			if err != nil {
				log.Fatal(err)
			}
			Info("Project updated successfully!")
		} else {
			Info("Project update canceled")
			return
		}
	}
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
		} else if text == "week" || text == "w" {
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
		} else if text == "week matrix" || text == "w m" {
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
		} else if text == "project list" || text == "projects" || text == "project" || text == "p" {
			projectMenu(TrackingRepositroy)
		} else if text == "project new" {
			clearTerminal()
			addProjectForm(TrackingRepositroy)
		} else if text == "exit" {
			break
		}
	}

}
