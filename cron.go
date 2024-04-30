package seoultechbot

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
)

var scheduler *gocron.Scheduler

func Cron(discord *discordgo.Session, Titles *formertitlelist) *gocron.Scheduler {
	Scheduler := gocron.NewScheduler(time.Local)
	Scheduler.SetMaxConcurrentJobs(3, gocron.WaitMode)
	Scheduler.Cron("0 0/1 * * *").Do(CheckUpdate, discord, AAI)
	Scheduler.Cron("0 0/1 * * *").Do(CheckUpdate, discord, COSS)
	Scheduler.Cron("0 0/1 * * *").Do(CheckUpdate, discord, SEOULTECH)
	return Scheduler
}

/*
func CheckTime() {
	loc, _ := time.LoadLocation("Asia/Seoul")
	fmt.Println("Scheduler works at ", time.Now().In(loc))
}
*/
