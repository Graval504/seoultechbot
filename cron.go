package seoultechbot

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
)

func Cron(discord *discordgo.Session) [3]*gocron.Job {
	Scheduler := gocron.NewScheduler(time.Local)
	job1, err := Scheduler.Every(1).Hour().Do(CheckUpdate, discord, AAI)
	if err != nil {
		fmt.Println(err)
	}
	job2, err := Scheduler.Every(1).Hour().Do(CheckUpdate, discord, COSS)
	if err != nil {
		fmt.Println(err)
	}
	job3, err := Scheduler.Every(1).Hour().Do(CheckUpdate, discord, SEOULTECH)
	if err != nil {
		fmt.Println(err)
	}
	return [3]*gocron.Job{job1, job2, job3}
}
