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
	Hours float64
	Ref   string
	Uid   string
	Date  time.Time
}

type Inprogress struct {
	Name      string
	Ref       string
	StartTime time.Time
}

type listFilter func(task Task) bool

func main() {
	var app = cli.NewApp()

	app.Name = "Quantum"
	app.Usage = "Simple time tracking"
	app.Description = "Simple command line application for tracking time spent on tasks"

	app.Commands = []cli.Command{
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list tasks",
			Subcommands: []cli.Command{
				{
					Name:     "month",
					Usage:    "List last months records",
					Action:   listMonthAction,
					Category: "Time filtering",
				},
				{
					Name:     "year",
					Usage:    "List last years records",
					Action:   listYearAction,
					Category: "Time filtering",
				},
				{
					Name:     "task",
					Usage:    "List tasks with matching task value",
					Action:   listTaskAction,
					Category: "Data filtering",
				},
				{
					Name:     "ref",
					Usage:    "List tasks with matching ref value",
					Action:   listRefAction,
					Category: "Data filtering",
				},
				{
					Name:     "inprogress",
					Usage:    "List in-progress tasks",
					Action:   listInprogressAction,
					Category: "Data filtering",
				},
			},
			Action: listDaysAction,
		},
		{
			Name:      "start",
			Usage:     "start a task",
			ArgsUsage: "task name",
			Action:    startAction,
		},
		{
			Name:      "stop",
			Usage:     "stop a task",
			ArgsUsage: "task name",
			Action:    stopAction,
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

func startAction(ctx *cli.Context) error {
	task := ctx.Args().First()
	if task == "" {
		fmt.Println("Incorrect usage of start \n")
		cli.ShowCommandHelpAndExit(ctx, "start", 1)
		return nil
	}

	db, err := openDb()
	if err != nil {
		return err
	}

	db.Write("inprogress", task, Inprogress{
		Name:      task,
		Ref:       ctx.Args().Get(1),
		StartTime: time.Now(),
	})

	return nil
}

func stopAction(ctx *cli.Context) error {
	task := ctx.Args().First()
	if task == "" {
		fmt.Println("Incorrect usage of stop \n")
		cli.ShowCommandHelpAndExit(ctx, "stop", 1)
		return nil
	}

	db, err := openDb()
	if err != nil {
		return err
	}

	inprogress := Inprogress{}
	if err := db.Read("inprogress", task, &inprogress); err != nil {
		return cli.NewExitError("Error reading database: "+err.Error(), 1)
	}

	uid := ksuid.New()

	db.Write("tasks", uid.String(), Task{
		Name:  task,
		Hours: time.Now().Sub(inprogress.StartTime).Hours(),
		Uid:   uid.String(),
		Ref:   inprogress.Ref,
		Date:  time.Now(),
	})

	if err := db.Delete("inprogress", task); err != nil {
		return cli.NewExitError("Error cleaning up inprogress task: "+err.Error(), 1)
	}

	return nil
}

func listInprogressAction(ctx *cli.Context) error {
	db, err := openDb()
	if err != nil {
		return err
	}

	records, err := db.ReadAll("inprogress")
	if err != nil {
		return cli.NewExitError("Error reading database: "+err.Error(), 1)
	}

	inprogress := [][]string{}
	for _, task := range records {
		inprogressFound := Inprogress{}
		if err := json.Unmarshal([]byte(task), &inprogressFound); err != nil {
			return cli.NewExitError("Error reading record: "+err.Error(), 1)
		}
		inprogress = append(inprogress, []string{inprogressFound.Name, inprogressFound.Ref, inprogressFound.StartTime.Format("2006-01-02 15:04:05")})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Task", "Ref", "Started"})
	table.AppendBulk(inprogress)
	table.Render()

	return nil
}

func listDaysAction(c *cli.Context) error {
	searchDays, error := strconv.Atoi(c.Args().First())
	if error != nil || searchDays == 0 {
		searchDays = 7
	}
	return listAction(c, func(task Task) bool {
		return task.Date.After(time.Now().AddDate(0, 0, -searchDays))
	})
}

func listMonthAction(c *cli.Context) error {
	return listAction(c, func(task Task) bool {
		return task.Date.After(time.Now().AddDate(0, -1, 0))
	})
}

func listYearAction(c *cli.Context) error {
	return listAction(c, func(task Task) bool {
		return task.Date.After(time.Now().AddDate(-1, 0, 0))
	})
}

func listTaskAction(c *cli.Context) error {
	return listAction(c, func(task Task) bool {
		return propertyMatches(c.Args(), task.Name)
	})
}

func listRefAction(c *cli.Context) error {
	return listAction(c, func(task Task) bool {
		return propertyMatches(c.Args(), task.Ref)
	})
}

func propertyMatches(args []string, property string) bool {
	for _, refArg := range args {
		if refArg == property {
			return true
		}
	}
	return false
}

func listAction(c *cli.Context, filter listFilter) error {
	db, err := openDb()
	if err != nil {
		return err
	}

	records, err := db.ReadAll("tasks")
	if err != nil {
		return cli.NewExitError("Error reading database: "+err.Error(), 1)
	}

	totalHours := 0.0
	tasks := [][]string{}
	for _, task := range records {
		taskFound := Task{}
		if err := json.Unmarshal([]byte(task), &taskFound); err != nil {
			return cli.NewExitError("Error reading record: "+err.Error(), 1)
		}

		if filter(taskFound) {
			tasks = append(tasks, []string{taskFound.Name, strconv.FormatFloat(taskFound.Hours, 'f', 2, 64), taskFound.Ref, taskFound.Date.Format("2006-01-02 15:04:05"), taskFound.Uid})
			totalHours += taskFound.Hours
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Task", "Hours", "Ref", "Date", "UID"})
	table.SetFooter([]string{"", "", "", "Total hours", strconv.FormatFloat(totalHours, 'f', 2, 64)})
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

	hoursFloat, err := strconv.ParseFloat(hours, 64)

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
		Hours: hoursFloat,
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
	if len(ctx.Args()) == 0 {
		fmt.Println("Incorrect usage of delete \n")
		cli.ShowCommandHelpAndExit(ctx, "delete", 1)
		return nil
	}

	db, err := openDb()
	if err != nil {
		return err
	}

	for _, uid := range ctx.Args() {
		if err := db.Delete("tasks", uid); err != nil {
			return cli.NewExitError("Error delete record from database: "+err.Error(), 1)
		}
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
