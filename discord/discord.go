package discord

import (
	"fmt"
	"nathanman/config"
	"nathanman/database"
	"nathanman/model"
	"nathanman/regex"
	"nathanman/util"
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
		util.PanicInRed(fmt.Sprintf("Bot couldn't be started! Token: %s. %s", config.Configuration.Discord.Token, err))
	}
	defer wg.Done()
	defer dg.Close()

	dg.AddHandler(ready)
	dg.AddHandler(messageCreate)

	if err = dg.Open(); err != nil {
		util.PanicInRed(fmt.Sprintf("Couldn't start bot! %v", err))
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
			util.LogInYellow(fmt.Sprintf("Couldn't get configuration for bot. %v", err))
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
				util.LogInYellow(fmt.Sprintf("Couldn't rename member %v to %v. %v", userName, entry.Name, err))
				return
			} else {
				config.Logger.Printf("Renamed user %s to %s! \n", userName, entry.Name)
			}
			if err := database.Db.Where("name = ?", entry.Name).Delete(&model.Entry{}).Error; err != nil {
				util.LogInYellow(fmt.Sprintf("Couldnt delete entry: %v. %v", entry, err))
				return
			}

			config_data.Lasttime = now.BeginningOfDay()
			if err := database.Db.Where("bot = ?", "nathanman").Save(config_data).Error; err != nil {
				util.LogInYellow(fmt.Sprintf("Couldn't save configuration: %v. %v", config_data, err))
			}
		}
	}
}

func setUsername(s *discordgo.Session) {
	tmp, usererror := s.User(config.Configuration.Discord.UserId)
	tmp2, usernickerror := s.GuildMember(config.Configuration.Discord.GuildId, config.Configuration.Discord.UserId)
	if usererror != nil {
		util.PanicInRed(fmt.Sprintf("UserId %s wasn't valid!, %s \n", config.Configuration.Discord.UserId, usererror))
	} else if usernickerror != nil {
		util.PanicInRed(fmt.Sprintf("UserId %s or GuildId %s wasn't valid!, %s \n", config.Configuration.Discord.UserId, config.Configuration.Discord.GuildId, usernickerror))
	}
	if tmp2.Nick != "" {
		userName = tmp2.Nick
	} else {
		userName = tmp.Username + "#" + tmp.Discriminator
	}

}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	if err := s.UpdateGameStatus(0, "Jonathan befreunden!"); err != nil {
		util.LogInYellow(fmt.Sprintf("Couldn't set game status. %v", err.Error()))
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
			if len(mp[0]) > 32 || regex.BadwordRegex.MatchString(mp[0]) {
				if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Nichtvalider nutzername: %v. %v", mp[0], m.Author.Mention())); err != nil {
					util.LogInYellow(fmt.Sprintf("Couldn't send plaintext message. %v", err))
				}
				return
			}

			// Execute only IF entry doesn't exist, aka there IS an error.
			if database.Db.Where("name = ?", mp[0]).First(&model.Entry{}).Error != nil {
				setUsername(s)
				entry := model.New(m.Author.ID, mp[0])
				database.Db.Create(entry)
				msg, err := s.ChannelMessageSendEmbed(config.Configuration.Discord.SendChannel, &discordgo.MessageEmbed{
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
					util.LogInYellow(fmt.Sprintf("Couldn't send embed message: %v. %v", msg, err))
				}
			}
		}
	}
}
