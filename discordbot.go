package seoultechbot

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func Discordbot(token string) {
	discord, err := discordgo.New("Bot " + token)
	//GetDiscordToken => Personal function that returns my disocrd bot Token. It doesn't exist on github.
	//So you should change this to your own bot Token
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	discord.AddHandler(messageCreate)
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}
	scheduler = Cron(discord, &TitleList)
	scheduler.StartAsync()
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	discord.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "addchannel",
			Description: "현재 채널을 공지 채널로 추가합니다.",
		},
		{
			Name:        "deletechannel",
			Description: "현재 채널을 공지 채널에서 제외합니다.",
		},
		{
			Name:        "checkchannel",
			Description: "현재 채널의 공지 알림 여부를 확인합니다.",
		},
		{
			Name:        "checkupdate",
			Description: "현재 업데이트 여부를 확인하여 공지합니다.",
		},
		{
			Name:        "savetitles",
			Description: "현재 게시글 목록을 서버에 저장합니다. *테스트용",
		},
		{
			Name:        "savechannels",
			Description: "알림을 받는 채널 목록을 서버에 저장합니다. *테스트용",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"addchannel":    AddChannel,
		"deletechannel": DeleteChannel,
		"checkupdate":   CheckUpdateNow,
		"savetitles":    SaveTitles,
		"savechannels":  SaveChannels,
		"checkchannel":  CheckChannel,
	}
)

func init() {
	flag.Parse()
	data, err := os.ReadFile("channelList.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
	}
	ChannelList = strings.Split(string(data), "\n")
}

func (b bulletin) SendUpdateInfo(discord *discordgo.Session, channelList []string) (errorList []error) {
	embed := &discordgo.MessageEmbed{
		Title: b.Title,
		URL:   b.Url,
		Color: 0x427eff,
		Image: &discordgo.MessageEmbedImage{
			URL: "attachment://image.png",
		},
	}

	if channelList == nil {
		return []error{errors.New("error: chnnel list is nil")}
	}
	c := make(chan error, len(channelList))

	for _, channel := range channelList {
		go SendEmbedImage(discord, embed, channel, b.Image, c)
	}
	for i := 0; i < len(channelList); i++ {
		err := <-c
		if err != nil {
			errorList = append(errorList, err)
		}
	}
	return errorList
}

func SendEmbed(discord *discordgo.Session, embed *discordgo.MessageEmbed, discordChannel string, c chan error) {
	_, err := discord.ChannelMessageSendEmbed(discordChannel, embed)
	if err != nil {
		c <- err
	}
	c <- nil
}

func SendEmbedImage(discord *discordgo.Session, embed *discordgo.MessageEmbed, discordChannel string, image []byte, c chan error) {
	imageReader := io.Reader(bytes.NewReader(image))
	_, err := discord.ChannelMessageSendComplex(
		discordChannel,
		&discordgo.MessageSend{
			Content: "",
			Files: []*discordgo.File{
				{
					Name:   "image.png",
					Reader: imageReader,
				},
			},
			Embed: embed,
		},
	)
	if err != nil {
		fmt.Println("error sending image:", err)
		c <- err
	}
	c <- err
}

var ChannelList = []string{}

func AddChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ChannelList = append(ChannelList, i.ChannelID)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: `<#` + i.ChannelID + `>` + "채널이 성공적으로 추가되었습니다.",
		},
	})
	os.WriteFile("channelList.txt", []byte(strings.Join(ChannelList, "\n")), 0644)
}

func CheckUpdateNow(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	var waitGroup sync.WaitGroup
	waitGroup.Add(3)
	for _, v := range [3]string{COSS, AAI, SEOULTECH} {
		go func(v string) {
			CheckUpdate(s, v)
			waitGroup.Done()
		}(v)
	}
	waitGroup.Wait()
	message := "업데이트되었습니다."
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &message,
	})
}

func CheckUpdate(s *discordgo.Session, url string) error {
	isUpdated, bulletinList, err := Scrap(url)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if !isUpdated {
		return nil
	}
	for _, v := range bulletinList {
		go func(v bulletin) {
			err := v.SendUpdateInfo(s, ChannelList)
			if len(err) != 0 {
				fmt.Println(err)
			}
		}(v)
	}
	return nil
}

func SaveTitles(s *discordgo.Session, i *discordgo.InteractionCreate) {
	TitleList.SaveFormerTitles()
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "업데이트되었습니다.",
		},
	})
}

func SaveChannels(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if err := os.WriteFile("channelList.txt", []byte(strings.Join(ChannelList, "\n")), 0644); err != nil {
		fmt.Println("error saving channelList,", err)
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "업데이트되었습니다.",
		},
	})
}

func DeleteChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	index := slices.IndexFunc(ChannelList, func(s string) bool { return s == i.ChannelID })
	if index == len(ChannelList) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: `<#` + i.ChannelID + `>` + "채널은 이미 알림을 받지 않는 상태입니다.",
			},
		})
		return
	}
	ChannelList = append(ChannelList[:index], ChannelList[index+1:]...)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: `<#` + i.ChannelID + `>` + "채널이 성공적으로 제거되었습니다.",
		},
	})
	os.WriteFile("channelList.txt", []byte(strings.Join(ChannelList, "\n")), 0644)
}

func CheckChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	index := slices.IndexFunc(ChannelList, func(s string) bool { return s == i.ChannelID })
	if index == len(ChannelList) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: `<#` + i.ChannelID + `>` + "채널은 알림을 받고 있는 상태입니다.",
			},
		})
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: `<#` + i.ChannelID + `>` + "채널은 알림을 받지 않는 상태입니다.",
		},
	})
}
