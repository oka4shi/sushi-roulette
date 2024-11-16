package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/exp/slices"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
)

var s *discordgo.Session

type channelInfo struct {
	session *discordgo.Session
	channel string
}

type SushiData []struct {
	Category string `json:"category"`
	Sushi    []struct {
		Name      string `json:"name"`
		ImagePath string `json:"img_url"`
	} `json:"sushi"`
}

type Sushi struct {
	Name      string
	ImagePath string
	Category  string
}

var shops = []string{"hama-sushi", "uobei"}

var data = map[string]*SushiData{}

func init() {
	token := os.Getenv("BOT_TOKEN")

	var errD error
	s, errD = discordgo.New("Bot " + token)
	if errD != nil {
		log.Fatalf("Invalid bot params: %v", errD)
	}

	for _, shop := range shops {
		file, err := os.Open(fmt.Sprintf("../dist/%s.json", shop))
		if err != nil {
			log.Fatal("Failed to read a json file: ", err)
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		data[shop] = &SushiData{}
		if err := decoder.Decode(data[shop]); err != nil {
			log.Fatal("Failed to parse a json file: ", err)
		}
	}
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "sushi-roulette",
			Description: "寿司ルーレット",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "brand",
					Description: "店の名前",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "はま寿司",
							Value: "hama-sushi",
						},
						{
							Name:  "魚べい",
							Value: "uobei",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "category",
					Description: "どんなメニューが出てくるか(デフォルト: 全て)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "寿司",
							Value: "sushi",
						},
						{
							Name:  "全て",
							Value: "normal",
						},
						{
							Name:  "デザート",
							Value: "dessert",
						},
						{
							Name:  "飲み物",
							Value: "drink",
						},
					},
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"sushi-roulette": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			brand := i.ApplicationCommandData().Options[0].StringValue()
			category := ""
			if len(i.ApplicationCommandData().Options) >= 2 {
				category = i.ApplicationCommandData().Options[1].StringValue()
			}

			var sushi []Sushi
			var nc []string
			switch brand {
			case "hama-sushi":
				switch category {
				case "sushi":
					nc = append(nc, "にぎり", "軍艦・細巻き・その他", "贅沢握り・三種盛り", "肉握り", "至福の一貫")
				case "dessert":
					nc = append(nc, "デザート・ドリンク")
				case "drink":
					nc = append(nc, "アルコール")
				default:
					nc = append(nc, "にぎり", "軍艦・細巻き・その他", "贅沢握り・三種盛り", "肉握り", "至福の一貫", "サイドメニュー", "期間限定")
				}
			case "uobei":
				switch category {
				case "sushi":
					nc = append(nc, "握り", "軍艦・巻物・いなり")
				case "dessert":
					nc = append(nc, "デザート・ドリンク")
				case "drink":
					nc = append(nc, "デザート・ドリンク")
				default:
					nc = append(nc, "握り", "軍艦・巻物・いなり", "サイドメニュー")

				}
			}

			for _, sushiAll := range *data[brand] {
				if slices.Contains(nc, sushiAll.Category) {
					for _, s := range sushiAll.Sushi {
						sushi = append(sushi, Sushi{s.Name, s.ImagePath, sushiAll.Category})
					}
				}
			}

			res := sushi[rand.IntN(len(sushi))]
			command_response_with_photo(s, i, fmt.Sprintf("%vにある%vのメニューから商品を選びました！\n\n%v\n%v", brand, category, res.Category, res.Name), res.ImagePath)
		},
	}
)

type errSendExec struct {
	ch  channelInfo
	err error
}

func (e errSendExec) sendMessage(message string) {
	if e.err == nil {
		e.ch.sendMessage(message)
	}
}

func (chinfo channelInfo) sendMessage(message string) {
	if strings.TrimSpace(message) == "" {
		//log.Println("Cannot send empty message")
		return
	}

	length := utf8.RuneCountInString(message)
	msgNum := length / 2000
	for i := 0; i < msgNum; i++ {
		_, err := chinfo.session.ChannelMessageSend(chinfo.channel, message[i*2000:(i+1)*2000+1])
		if err != nil {
			log.Printf("Cannnot send a message: %v", err)
		}
	}

	_, err := chinfo.session.ChannelMessageSend(chinfo.channel, message[msgNum*2000:])
	if err != nil {
		log.Printf("Cannnot send a message: %v", err)
	}
}

func command_response(se *discordgo.Session, it *discordgo.InteractionCreate, message string) string {
	err := se.InteractionRespond(it.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
	if err != nil {
		log.Printf("Cannot respond to command: %v", err)
	}

	mes, err := se.InteractionResponse(it.Interaction)
	if err != nil {
		log.Printf("Cannot get response: %v", err)
	}
	return mes.ID
}

func command_response_with_photo(se *discordgo.Session, it *discordgo.InteractionCreate, message string, imageUrl string) string {
	err := se.InteractionRespond(it.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Embeds: []*discordgo.MessageEmbed{{
				Image: &discordgo.MessageEmbedImage{URL: imageUrl},
			}},
		},
	})
	if err != nil {
		log.Printf("Cannot respond to command: %v", err)
	}

	mes, err := se.InteractionResponse(it.Interaction)
	if err != nil {
		log.Printf("Cannot get response: %v", err)
	}
	return mes.ID
}

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session %v", err)
	}

	log.Println("adding commands")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	commands, err := s.ApplicationCommands(s.State.User.ID, "")
	for i := range commands {
		command := commands[i]
		log.Printf("%v", command)
		err := s.ApplicationCommandDelete(s.State.User.ID, "", command.ID)
		if err != nil {
			log.Printf("Cannot delete command: %v", err)
		} else {
			log.Printf("%v was deleted!", command.Name)

		}
	}

	log.Println("exit")

}
