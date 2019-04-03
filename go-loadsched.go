package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "go-loadsched"
	app.Version = "0.1.0"
	app.Usage = "Load shedding schedule tool."

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "day, d",
			Value: -1,
			Usage: "Day of month to show schedule for.",
		},
		cli.StringFlag{
			Name:  "filename, f",
			Value: "schedule.txt",
			Usage: "File to read schedule from.",
		},
		cli.IntSliceFlag{
			Name:  "group, g",
			Usage: "Filter on specified groups.",
		},
		cli.IntFlag{
			Name:  "stage, s",
			Value: -1,
			Usage: "Schedule for this stage only. If not given, it is queried from Eskom.",
		},
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "Print verbose output.",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "print-schedule",
			Usage: "Print the entire schedule.",
			Action: func(c *cli.Context) error {
				schedule := Schedule{FileName: c.Parent().String("filename")}
				err := schedule.Load()
				if err != nil {
					return err
				}

				schedule.Print()
				return nil
			},
		},
	}
	app.Action = func(c *cli.Context) {
		verbose := c.Bool("verbose")

		schedule := &Schedule{FileName: c.String("filename")}
		err := schedule.Load()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		stage := c.Int("stage")
		if stage < 0 {
			stage, err = fetchStage()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if stage == 0 {
				if verbose {
					fmt.Println("# No load shedding in progress! \\o/")
				}
				return
			}
		}
		schedule = schedule.FilterByStage(stage)
		if verbose {
			fmt.Println("# LOAD SHEDDING STAGE:", stage)
		}

		dayOfMonth := c.Int("day")
		if dayOfMonth < 0 {
			dayOfMonth = fetchToday()
		}
		if verbose {
			fmt.Println("# DAY OF MONTH:", dayOfMonth)
		}
		schedule = schedule.FilterByDay(dayOfMonth)

		groups := c.IntSlice("group")
		if verbose {
			var groupStrs []string
			for _, g := range groups {
				groupStrs = append(groupStrs, strconv.Itoa(g))
			}
			fmt.Println("# GROUPS:", strings.Join(groupStrs, ", "))
		}
		if len(groups) > 0 {
			schedule = schedule.FilterByGroups(groups)
		}

		if verbose {
			fmt.Println("# TIMESLOTS:")
		}
		schedule.Print()
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func fetchStage() (int, error) {
	resp, err := http.Get("http://loadshedding.eskom.co.za/LoadShedding/getstatus")
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}

	statusId, err := strconv.Atoi(string(respBody))

	if err != nil {
		return -1, err
	}

	if statusId == 99 {
		return -1, errors.New("Eskom data is unavailable")
	}

	return statusId - 1, nil
}

func fetchToday() int {
	return time.Now().Day()
}
