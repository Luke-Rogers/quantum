package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/nanobox-io/golang-scribble"
	"github.com/olekukonko/tablewriter"
	"github.com/segmentio/ksuid"
	"gopkg.in/urfave/cli.v1"
)

type Task struct {
	Name  string
	Hours int
	Ref   string
	Uid   string
	Date  time.Time
}

func main() {
	var app = cli.NewApp()

	app.Name = "Quantum"
	app.Usage = "Simple time tracking"
	app.Description = "Simple command line application for tracking time spend on tasks"

	app.Commands = []cli.Command{
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list tasks",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "days",
					Usage: "Number of days back to run the search, default is 7",
				},
			},
			Action: listAction,
		},
		{
			Name:      "add",
			Aliases:   []string{"a"},
			Usage:     "add a task",
			ArgsUsage: "TASK (Mandatory) HOURS (Mandatory) REF (Optional)",
			Action:    addAction,
		},
		{
			Name:      "delete",
			Aliases:   []string{"d"},
			Usage:     "delete a task by uid or all",
			ArgsUsage: "UID",
			Subcommands: []cli.Command{
				{
					Name:   "all",
					Usage:  "Delete all records",
					Action: deleteAllAction,
				},
			},
			Action: deleteAction,
		},
	}

	app.Run(os.Args)
}

func listAction(c *cli.Context) error {
	db, err := openDb()
	if err != nil {
		return err
	}

	records, err := db.ReadAll("tasks")
	if err != nil {
		return cli.NewExitError("Error reading database: "+err.Error(), 1)
	}

	searchDays := c.Int("days")
	if searchDays == 0 {
		searchDays = 7
	}
	afterDate := time.Now().AddDate(0, 0, -searchDays).Truncate(time.Hour)
	fmt.Println(afterDate)
	totalHours := 0
	tasks := [][]string{}
	for _, task := range records {
		taskFound := Task{}
		if err := json.Unmarshal([]byte(task), &taskFound); err != nil {
			return cli.NewExitError("Error reading record: "+err.Error(), 1)
		}

		if taskFound.Date.After(afterDate) {
			tasks = append(tasks, []string{taskFound.Name, strconv.Itoa(taskFound.Hours), taskFound.Ref, taskFound.Date.Format("2006-01-02 15:04:05"), taskFound.Uid})
			totalHours += taskFound.Hours
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Task", "Hours", "Ref", "Date", "UID"})
	table.SetFooter([]string{"", "", "", "Total hours", strconv.Itoa(totalHours)})
	table.AppendBulk(tasks)
	table.Render()

	return nil
}

func addAction(ctx *cli.Context) error {
	task := ctx.Args().Get(0)
	hours := ctx.Args().Get(1)
	ref := ctx.Args().Get(2)

	if task == "" || hours == "" {
		fmt.Println("Incorrect usage of add \n")
		cli.ShowCommandHelpAndExit(ctx, "add", 1)
		return nil
	}

	hourInt, err := strconv.Atoi(hours)

	if err != nil {
		fmt.Println("Incorrect usage of add \n")
		cli.ShowCommandHelpAndExit(ctx, "add", 1)
		return nil
	}

	db, err := openDb()
	if err != nil {
		return err
	}

	uid := ksuid.New()

	db.Write("tasks", uid.String(), Task{
		Name:  task,
		Hours: hourInt,
		Uid:   uid.String(),
		Ref:   ref,
		Date:  time.Now(),
	})

	return nil
}

func deleteAllAction(ctx *cli.Context) error {
	db, err := openDb()
	if err != nil {
		return err
	}
	if err := db.Delete("tasks", ""); err != nil {
		return cli.NewExitError("Error deleting all records from database: "+err.Error(), 1)
	}
	return nil
}

func deleteAction(ctx *cli.Context) error {
	uid := ctx.Args().Get(0)
	if uid == "" {
		fmt.Println("Incorrect usage of delete \n")
		cli.ShowCommandHelpAndExit(ctx, "delete", 1)
		return nil
	}

	db, err := openDb()
	if err != nil {
		return err
	}

	if err := db.Delete("tasks", uid); err != nil {
		return cli.NewExitError("Error delete record from database: "+err.Error(), 1)
	}

	return nil
}

func openDb() (*scribble.Driver, error) {
	userHomeDir, error := homedir.Dir()
	if error != nil {
		return nil, cli.NewExitError("Unable to resolve user home dir", 1)
	}
	dbDir := userHomeDir + "/.quantum"
	db, err := scribble.New(dbDir, nil)
	if err != nil {
		return nil, cli.NewExitError("Unable to open database: "+dbDir, 1)
	}
	return db, nil
}