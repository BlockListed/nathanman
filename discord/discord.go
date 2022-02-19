package discord

import (
	"fmt"
	"nathanman/config"
	"nathanman/database"
	"nathanman/model"
	"nathanman/regex"
	"sync"
	"time"

	"github.com/TwiN/go-color"
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
		config.Logger.Panicln(color.InRed(fmt.Sprintf("Bot couldn't be started! Token: %s. %s", config.Configuration.Discord.Token, err)))
	}
	defer wg.Done()
	defer dg.Close()

	dg.AddHandler(ready)
	dg.AddHandler(messageCreate)

	if err = dg.Open(); err != nil {
		config.Logger.Panicln(color.InRed(fmt.Sprintf("Couldn't start bot! %v", err)))
	}
	go poll_rename(dg)

	<-quit
}

func poll_rename(s *discordgo.Session) {
	var config_data model.Config

	duration, _ := time.ParseDuration("1m")

	for {
		time.Sleep(duration)
		if err := database.Db.Where("bot = ?", "nathanman").Take(&config_data).Error; err != nil {
			config.Logger.Println(color.InYellow(fmt.Sprintf("Couldn't get configuration for bot. %v", err)))
			continue
		}
		nextDayLastime := now.With(config_data.Lasttime).BeginningOfDay().AddDate(0, 0, 1)
		nowtime := time.Now()
		if nowtime.After(nextDayLastime) || nowtime.Equal(nextDayLastime) {
			var entry model.Entry
			err := database.Db.Order("RANDOM ()").First(&entry).Error
			if err != nil {
				entry.Name = "HabKeinenNamenFürDenKeknathan"
			}

			setUsername(s)
			if err := s.GuildMemberNickname(config.Configuration.Discord.GuildId, config.Configuration.Discord.UserId, entry.Name); err != nil {
				config.Logger.Println(color.InYellow(fmt.Sprintf("Couldn't rename member %v to %v. %v", userName, entry.Name, err)))
				return
			} else {
				config.Logger.Printf("Renamed user %s to %s! \n", userName, entry.Name)
			}
			if err := database.Db.Where("name = ?", entry.Name).Delete(&model.Entry{}).Error; err != nil {
				config.Logger.Println(color.InYellow(fmt.Sprintf("Couldnt delete entry: %v. %v", entry, err)))
				return
			}

			config_data.Lasttime = now.BeginningOfDay()
			if err := database.Db.Where("bot = ?", "nathanman").Save(config_data).Error; err != nil {
				config.Logger.Println(color.InYellow(fmt.Sprintf("Couldn't save configuration: %v. %v", config_data, err)))
			}
		}
	}
}

func setUsername(s *discordgo.Session) {
	tmp, usererror := s.User(config.Configuration.Discord.UserId)
	tmp2, usernickerror := s.GuildMember(config.Configuration.Discord.GuildId, config.Configuration.Discord.UserId)
	if usererror != nil {
		config.Logger.Panicln(color.InRed(fmt.Sprintf("UserId %s wasn't valid!, %s \n", config.Configuration.Discord.UserId, usererror)))
	} else if usernickerror != nil {
		config.Logger.Panicln(color.InRed(fmt.Sprintf("UserId %s or GuildId %s wasn't valid!, %s \n", config.Configuration.Discord.UserId, config.Configuration.Discord.GuildId, usernickerror)))
	}
	if tmp2.Nick != "" {
		userName = tmp2.Nick
	} else {
		userName = tmp.Username + "#" + tmp.Discriminator
	}

}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	if err := s.UpdateGameStatus(0, "Jonathan befreunden!"); err != nil {
		config.Logger.Println(color.InYellow(fmt.Sprintf("Couldn't set game status. %v", err.Error())))
	}
	setUsername(s)
	config.Logger.Printf("Bot started as %s.", s.State.User.Username)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.ChannelID == config.Configuration.Discord.ListenChannel {
		if mp := regex.Regex.FindStringSubmatch(m.Content); mp != nil {
			name := mp[0]
			if len(name) > 32 || regex.BadwordRegex.MatchString(name) {
				msg, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
					Content: fmt.Sprintf("Nichtvalider nutzername: %v. %v", name, m.Author.Mention()),
					Reference: &discordgo.MessageReference{
						MessageID: m.ID,
						ChannelID: m.ChannelID,
						GuildID:   m.GuildID,
					},
				})
				if err != nil {
					config.Logger.Println(color.InYellow(fmt.Sprintf("Couldn't send complex message message: %v. %v", msg, err)))
				}
				config.Logger.Println(color.InYellow(fmt.Sprintf("Invalid name: %v", name)))
				return
			}

			// Execute only IF entry doesn't exist, aka there IS an error.
			if database.Db.Where("name = ?", name).First(&model.Entry{}).Error != nil {
				setUsername(s)
				entry := model.New(m.Author.ID, name)
				database.Db.Create(entry)
				msg, err := s.ChannelMessageSendEmbed(config.Configuration.Discord.SendChannel, &discordgo.MessageEmbed{
					Type:  discordgo.EmbedTypeRich,
					Title: "Nutzername für " + userName + " hinzugefügt. Name: " + name,
					Color: 32768,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Nutzername ID: ",
							Value: entry.ID,
						},
					},
				})
				if err != nil {
					config.Logger.Println(color.InYellow(fmt.Sprintf("Couldn't send embed message: %v. %v", msg, err)))
				}
			}
		}
	}
}
