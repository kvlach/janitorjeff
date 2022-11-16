package testkit

// func GetRandomMessages(n int) []*TestMessage {
// 	if n <= 0 {
// 		panic("unexpected n")
// 	}

// 	messages := []*TestMessage{}

// 	for i := 0; i < n; i++ {
// 		tm := NewTestMessage().DiscordRandom()
// 		messages = append(messages, tm)
// 	}

// 	return messages
// }

// type DiscordMessage struct {
// 	ID     string `json:"id"`
// 	Text   string `json:"text"`
// 	IsDm   bool   `json:"is_dm"`
// 	Author struct {
// 		ID          string `json:"id"`
// 		Name        string `json:"name"`
// 		DisplayName string `json:"display_name"`
// 		Mention     string `json:"mention"`
// 	} `json:"author"`
// 	Channel struct {
// 		ID   string `json:"id"`
// 		Name string `json:"name"`
// 	} `json:"channel"`
// }

// func GetDiscordMessages() ([]*core.Message, error) {
// 	data, err := ioutil.ReadFile("../../testkit/data/discord.json")
// 	if err != nil {
// 		return nil, err
// 	}

// 	var discordMsgs []DiscordMessage
// 	err = json.Unmarshal(data, &discordMsgs)

// 	var msgs []*core.Message
// 	for _, dmsg := range discordMsgs {
// 		// TODO: display name
// 		dgMsg := &dg.MessageCreate{
// 			Message: &dg.Message{
// 				ID:      dmsg.ID,
// 				Content: dmsg.Text,
// 				Author: &dg.User{
// 					ID:       dmsg.Author.ID,
// 					Username: dmsg.Author.Name,
// 				},
// 				ChannelID: dmsg.Channel.ID,
// 				GuildID:   "", // isDM
// 			},
// 		}

// 		msgTmp := discord.DiscordMessageCreate{Session: nil, Message: dgMsg}

// 		msg, err := msgTmp.Parse()
// 		if err != nil {
// 			return nil, err
// 		}
// 		msgs = append(msgs, msg)
// 	}

// 	return msgs, err
// }

// func (tdb *TestDB) PopulateDB() *TestDB {
// 	msgs, err := GetDiscordMessages()
// 	if err != nil {
// 		log.Fatalf("%v\n", err)
// 	}

// 	for _, m := range msgs {
// 		if _, err := m.Scope(); err != nil {
// 			log.Fatalf("%v\n", err)
// 		}

// 		if _, err := m.Scope(discord.Author); err != nil {
// 			log.Fatalf("%v\n", err)
// 		}
// 	}

// 	return tdb
// }
