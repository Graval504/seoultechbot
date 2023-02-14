package seoultechbot

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
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
			Name:        "checkupdate",
			Description: "현재 업데이트 여부를 확인하여 공지합니다.",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"addchannel":  AddChannel,
		"checkupdate": CheckUpdate,
	}
)

func init() {
	flag.Parse()
}

func (b bulletin) SendUpdateInfo(discord *discordgo.Session, channelList []string) (errorList []error) {
	embed := &discordgo.MessageEmbed{
		Title: b.title,
		URL:   b.url,
		Image: &discordgo.MessageEmbedImage{
			URL: b.image,
		},
		Color: 0x427eff,
	}
	if channelList == nil {
		return []error{errors.New("error: chnnel list is nil")}
	}
	c := make(chan error, len(channelList))
	for _, channel := range channelList {
		go SendEmbed(discord, embed, channel, c)
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

var ChannelList = []string{}

func AddChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ChannelList = append(ChannelList, i.ChannelID)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: `<#` + i.ChannelID + `>` + "가 성공적으로 추가되었습니다.",
		},
	})
}

func CheckUpdate(s *discordgo.Session, i *discordgo.InteractionCreate) {

}
