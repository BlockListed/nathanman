package discord

import (
	"nathanman/config"
	"nathanman/database"
	"nathanman/model"
	"nathanman/regex"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jinzhu/now"
)

var userName string

func Run(quit chan interface{}, wg *sync.WaitGroup) {
	if config.Configuration.Discord.Token == "" {
		config.Logger.Panic("Invalid Token provided!\n")
	}

	dg, err := discordgo.New("Bot " + config.Configuration.Discord.Token)
	if err != nil {
		config.Logger.Panicf("Bot couldn't be started! Token: %s\n", config.Configuration.Discord.Token)
	}
	defer wg.Done()
	defer dg.Close()

	dg.AddHandler(ready)
	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		config.Logger.Panicln("Couldn't start bot!")
	}
	go poll_rename(dg)

	<-quit
}

func poll_rename(s *discordgo.Session) {
	var config_data model.Config

	duration, _ := time.ParseDuration("1m")

	for {
		database.Db.Where("bot = ?", "nathanman").Take(&config_data)
		nextDayLastime := now.With(config_data.Lasttime).BeginningOfDay().AddDate(0, 0, 1)
		nowtime := time.Now()
		if nowtime.After(nextDayLastime) || nowtime.Equal(nextDayLastime) {
			var entry model.Entry
			err := database.Db.Order("RANDOM ()").First(&entry).Error
			if err != nil {
				entry.Name = "HabKeinenNamenFürDenKeknathan"
			}

			s.GuildMemberNickname(config.Configuration.Discord.GuildId, config.Configuration.Discord.UserId, entry.Name)
			config.Logger.Printf("Renamed user %s to %s! \n", userName, entry.Name)
			database.Db.Where("name = ?", entry.Name).Delete(&model.Entry{})

			config_data.Lasttime = now.BeginningOfDay()
			database.Db.Where("bot = ?", "nathanman").Save(config_data)
		}
		time.Sleep(duration)
	}
}

func setUsername(s *discordgo.Session) {
	tmp, err := s.User(config.Configuration.Discord.UserId)
	tmp2, err2 := s.GuildMember(config.Configuration.Discord.GuildId, config.Configuration.Discord.UserId)
	if err != nil {
		config.Logger.Panicf("UserId %s wasn't valid!, %s \n", config.Configuration.Discord.UserId, err)
	} else if err != nil {
		config.Logger.Panicf("UserId %s or GuildId %s wasn't valid!, %s \n", config.Configuration.Discord.UserId, config.Configuration.Discord.GuildId, err2)
	}
	if tmp2.Nick != "" {
		userName = tmp2.Nick
	} else {
		userName = tmp.Username + "#" + tmp.Discriminator
	}

}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "Jonathan befreunden!")
	setUsername(s)
	config.Logger.Printf("Bot started as %s.", s.State.User.Username)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.ChannelID == config.Configuration.Discord.ListenChannel {
		if mp := regex.Regex.FindStringSubmatch(m.Content); mp != nil {
			if database.Db.Where("name = ?", mp[0]).First(&model.Entry{}).Error != nil {
				setUsername(s)
				entry := model.New(m.Author.ID, mp[0])
				database.Db.Create(entry)
				_, err := s.ChannelMessageSendEmbed(config.Configuration.Discord.SendChannel, &discordgo.MessageEmbed{
					Type:  discordgo.EmbedTypeRich,
					Title: "Nutzername für " + userName + " hinzugefügt. Name: " + mp[0],
					Color: 32768,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Nutzername ID: ",
							Value: entry.ID,
						},
					},
				})
				if err != nil {
					config.Logger.Printf("\033[31m Couldn't send message %v \033[0m", err)
				}
			}
		}
	}
}
