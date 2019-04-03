package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var SET = struct{}{}

type TimeSlot struct {
	start, end string
}

type ScheduleData map[int]map[TimeSlot]map[int]map[int]struct{}

type Schedule struct {
	FileName string
	Data     ScheduleData
}

type ScheduleIface interface {
	Load() error
	FilterByDay(int) *Schedule
	FilterByGroups([]int) *Schedule
	FilterByStage(int) *Schedule
	Print()
}

func (schedule *Schedule) Load() error {
	f, err := os.Open(schedule.FileName)
	if err != nil {
		return err
	}

	defer f.Close()

	fileData, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	schedule.Data, err = parseScheduleFile(string(fileData[:]))
	if err != nil {
		return err
	}

	return nil
}

func (schedule *Schedule) FilterByDay(dayOfMonth int) *Schedule {
	filtered := make(ScheduleData)
	filtered[dayOfMonth] = schedule.Data[dayOfMonth]
	return &Schedule{FileName: schedule.FileName, Data: filtered}
}

func (schedule *Schedule) FilterByGroups(groups []int) *Schedule {
	filtered := make(ScheduleData)

	for dayOfMonth, timeSlots := range schedule.Data {
		for timeSlot, stages := range timeSlots {
			for stage, stageGroups := range stages {
				for _, group := range groups {
					_, ok := stageGroups[group]
					if !ok {
						continue
					}

					if filtered[dayOfMonth] == nil {
						filtered[dayOfMonth] = make(map[TimeSlot]map[int]map[int]struct{})
					}
					if filtered[dayOfMonth][timeSlot] == nil {
						filtered[dayOfMonth][timeSlot] = make(map[int]map[int]struct{})
					}
					if filtered[dayOfMonth][timeSlot][stage] == nil {
						filtered[dayOfMonth][timeSlot][stage] = make(map[int]struct{})
					}

					filtered[dayOfMonth][timeSlot][stage][group] = SET
				}
			}
		}
	}

	return &Schedule{FileName: schedule.FileName, Data: filtered}
}

func (schedule *Schedule) FilterByStage(stage int) *Schedule {
	filtered := make(ScheduleData)

	for dayOfMonth, timeSlots := range schedule.Data {
		for timeSlot, stages := range timeSlots {
			if filtered[dayOfMonth] == nil {
				filtered[dayOfMonth] = make(map[TimeSlot]map[int]map[int]struct{})
			}
			if filtered[dayOfMonth][timeSlot] == nil {
				filtered[dayOfMonth][timeSlot] = make(map[int]map[int]struct{})
			}
			filtered[dayOfMonth][timeSlot][stage] = stages[stage]
		}
	}

	return &Schedule{FileName: schedule.FileName, Data: filtered}
}

func (schedule *Schedule) Print() {
	var dayKeys []int
	for k := range schedule.Data {
		dayKeys = append(dayKeys, k)
	}
	sort.Ints(dayKeys)

	for _, dayOfMonth := range dayKeys {
		timeSlots := schedule.Data[dayOfMonth]

		var timeSlotKeys []TimeSlot
		for ts, _ := range timeSlots {
			timeSlotKeys = append(timeSlotKeys, ts)
		}
		sort.Slice(timeSlotKeys, func(i, j int) bool {
			return TimeToInt(timeSlotKeys[i].start) < TimeToInt(timeSlotKeys[j].start)
		})

		for _, timeSlot := range timeSlotKeys {
			stages := timeSlots[timeSlot]

			var stageKeys []int
			for k, _ := range stages {
				stageKeys = append(stageKeys, k)
			}
			sort.Ints(stageKeys)

			for _, stage := range stageKeys {
				groups := stages[stage]

				if len(groups) == 0 {
					continue
				}

				var groupNums []int
				for gnum := range groups {
					groupNums = append(groupNums, gnum)
				}
				sort.Ints(groupNums)

				var groupStrs []string
				for _, g := range groupNums {
					groupStrs = append(groupStrs, strconv.Itoa(g))
				}

				fmt.Printf(
					"%2d %d %5s - %5s: %s\n",
					dayOfMonth, stage, timeSlot.start, timeSlot.end,
					strings.Join(groupStrs, ", "),
				)
			}
		}
	}
}

func parseScheduleFile(contents string) (ScheduleData, error) {
	var time TimeSlot
	data := make(ScheduleData)

	for _, line := range strings.Split(contents, "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}

		fields := RemoveEmpty(regexp.MustCompile("\\s*\\|\\s*").Split(line, -1))

		if len(fields) == 2 {
			time = TimeSlot{start: fields[0], end: fields[1]}
		} else if len(fields) == 32 {
			stageStr := fields[0]
			fields = fields[1:]

			stage, err := strconv.Atoi(strings.Replace(stageStr, "Stage", "", -1))
			if err != nil {
				return nil, err
			}

			for i, field := range fields {
				dayOfMonth := i + 1

				groupNum, err := strconv.Atoi(field)
				if err != nil {
					return nil, err
				}

				if data[dayOfMonth] == nil {
					data[dayOfMonth] = make(map[TimeSlot]map[int]map[int]struct{})
				}
				if data[dayOfMonth][time] == nil {
					data[dayOfMonth][time] = make(map[int]map[int]struct{})
				}
				if data[dayOfMonth][time][stage] == nil {
					data[dayOfMonth][time][stage] = make(map[int]struct{})
				}

				timeSlot := data[dayOfMonth][time]
				if timeSlot[stage] == nil {
					timeSlot[stage] = make(map[int]struct{})
				}
				timeSlot[stage][groupNum] = SET

				if stage > 1 {
					for gn := range timeSlot[stage-1] {
						timeSlot[stage][gn] = SET
					}
				}
			}
		}
	}

	return data, nil
}
