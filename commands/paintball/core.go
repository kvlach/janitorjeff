package paintball

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

// designed for discord, not with cross-platform in mind

const (
	assetFakePoster = "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dc/F_for_Fake_%281973_poster%29.jpg/1200px-F_for_Fake_%281973_poster%29.jpg"
	assetQuestion   = "https://media2.giphy.com/media/3FogJGpt7jfu5zlKdB/giphy.gif"

	categoryPoster = iota
	categoryScramble
	categoryFakeOrReal
	categoryYear
	categoryDirector
	categoryTrueOrFalse
	categoryPlot

	embedColor = 0xD0021A

	interval = 5 * time.Second
)

var categories = []int{
	categoryPoster,
	categoryScramble,
	categoryFakeOrReal,
	categoryYear,
	categoryDirector,
	// categoryTrueOrFalse,
	categoryPlot,
}

var (
	game       = createGame()
	movies     = readMovies()
	fakeMovies = readFakeMovies()
)

// don't want to quote reply since it always quotes the original command call
// message and since this command is discord only it's fine to use this
func write(channel string, embed *dg.MessageEmbed) {
	embed.Color = embedColor
	discord.Session.ChannelMessageSendEmbed(channel, embed)
}

type score struct {
	player string
	points int
}

type pb struct {
	sync.RWMutex
	active map[int64][]*score
}

func createGame() *pb {
	return (&pb{}).Init()
}

func (pb *pb) Init() *pb {
	pb.Lock()
	defer pb.Unlock()

	if pb.active == nil {
		pb.active = make(map[int64][]*score)
	}
	return pb
}

func (pb *pb) Active(place int64) bool {
	pb.RLock()
	defer pb.RUnlock()

	_, ok := pb.active[place]
	return ok
}

func (pb *pb) Playing(place int64, v bool) {
	pb.Lock()
	defer pb.Unlock()

	if v == true {
		pb.active[place] = []*score{}
	} else {
		delete(pb.active, place)
	}
}

func (pb *pb) Point(place int64, player *core.User) {
	pb.Lock()
	defer pb.Unlock()

	// Discord mention use's the users's ID, which means that it is static and
	// so an ok way to keep track of score

	for _, s := range pb.active[place] {
		if s.player == player.Mention {
			s.points += 1
			return
		}
	}
	pb.active[place] = append(pb.active[place], &score{player.Mention, 1})
}

func (pb *pb) Scores(place int64) []*score {
	pb.RLock()
	defer pb.RUnlock()

	return pb.active[place]
}

type movie struct {
	Title     string   `json:"title"`
	Year      int      `json:"year"`
	Directors []string `json:"directors"`
	Poster    string   `json:"poster"`
	Plot      string   `json:"plot"`
}

func readMovies() []movie {
	content, err := ioutil.ReadFile("data/movies.json")
	if err != nil {
		panic(err)
	}

	var movies []movie
	err = json.Unmarshal(content, &movies)
	if err != nil {
		panic(err)
	}

	return movies
}

func readFakeMovies() []string {
	file, err := os.Open("data/fake-movies.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return lines
}

func randomMovie() movie {
	rand.Seed(time.Now().UnixNano())
	return movies[rand.Intn(len(movies))]
}

func randomFakeMovie() string {
	rand.Seed(time.Now().UnixNano())
	return fakeMovies[rand.Intn(len(fakeMovies))]
}

func shuffle(s string) string {
	re := regexp.MustCompile(`[^\s]+`)
	return re.ReplaceAllStringFunc(s, func(m string) string {
		rand.Seed(time.Now().UnixNano())
		runes := []rune(m)
		rand.Shuffle(len(m), func(i, j int) {
			runes[i], runes[j] = runes[j], runes[i]
		})
		return strings.ToLower(string(runes))
	})
}

func simplify(s string) string {
	re := regexp.MustCompile(`[^\w]`)
	return strings.ToLower(re.ReplaceAllString(s, ""))
}

func awaitAnswer(here int64, answers []string) *core.Message {
	msg := core.Await(15*time.Second, func(m *core.Message) bool {
		place, err := m.HereExact()
		if err != nil {
			return false
		}

		if place != here {
			return false
		}

		for _, a := range answers {
			if simplify(a) == simplify(m.Raw) {
				return true
			}
		}
		return false
	})
	return msg
}

func generateQuestion(round int) (*dg.MessageEmbed, []string, string) {
	rand.Seed(time.Now().UnixNano())
	category := categories[rand.Intn(len(categories))]

	var desc strings.Builder
	var question string
	var icon string
	var poster string
	var answers []string
	var image string

	m := randomMovie()

	switch category {
	case categoryPoster:
		icon = "🖼"
		question = "Name the movie from the POSTER:"
		answers = append(answers, m.Title)
		image = m.Poster

	case categoryScramble:
		icon = "🧩"
		question = "UNSCRAMBLE the movie:"
		answers = append(answers, m.Title)
		fmt.Fprintf(&desc, "*%s*\n\n", shuffle(m.Title))

	case categoryFakeOrReal:
		icon = "🔍"
		question = "IS this movie FAKE or REAL?"

		var title string
		switch rand.Intn(2) {
		case 0:
			title = m.Title
			answers = append(answers, "Real")
		case 1:
			title = randomFakeMovie()
			answers = append(answers, "Fake")
			poster = assetFakePoster
		}
		fmt.Fprintf(&desc, "%s\n\n", title)

	case categoryYear:
		icon = "📆"
		question = "What YEAR was this movie released in?"
		answers = append(answers, fmt.Sprint(m.Year))
		fmt.Fprintf(&desc, "%s\n\n", m.Title)

	case categoryDirector:
		icon = "📣"
		question = "Who DIRECTED this movie?"
		answers = append(answers, m.Directors...)
		fmt.Fprintf(&desc, "%s (%d)\n\n", m.Title, m.Year)

	case categoryTrueOrFalse:
		icon = "🤔"
		question = "Is this statement TRUE or FALSE?"

	case categoryPlot:
		icon = "📖"
		question = "Name the movie from the PLOT:"
		answers = append(answers, m.Title)
		fmt.Fprintf(&desc, "*%s*\n\n", m.Plot)
	}

	desc.WriteString("Enter your answer in the chat!\n")
	if category == categoryPoster {
		desc.WriteString("\n")
	}
	desc.WriteString("You have 15 seconds.\n")

	if poster == "" {
		// serve a smaller image instead of the full res one since the
		// thumbnail image in which they are put is quite small
		poster = m.Poster + "._V1_SX300.jpg"
	}

	embed := &dg.MessageEmbed{
		Title:       fmt.Sprintf("%s Round %d: %s", icon, round, question),
		Description: desc.String(),
		Image: &dg.MessageEmbedImage{
			URL: image,
		},
		Thumbnail: &dg.MessageEmbedThumbnail{
			URL: assetQuestion,
		},
	}

	return embed, answers, poster
}

func generateAnswer(round int, winner, poster string, answers []string, last bool) *dg.MessageEmbed {
	var title string
	if winner == "" {
		title = fmt.Sprintf("**Round %d: Nobody Answered**", round)
	} else {
		title = fmt.Sprintf("**Round %d: %s got the Answer!**", round, winner)
	}

	var desc strings.Builder
	answer := strings.Join(answers, " **or** ")
	fmt.Fprintf(&desc, "The correct answer was: *%s*\n", answer)

	if winner != "" {
		desc.WriteString("**1 point**\n")
	}

	if !last {
		desc.WriteString("\nNext Question in a few seconds!\n")
	}

	embed := &dg.MessageEmbed{
		Title:       title,
		Description: desc.String(),
		Thumbnail: &dg.MessageEmbedThumbnail{
			URL: poster,
		},
	}

	return embed
}

func generateScorecard(scores []*score) *dg.MessageEmbed {
	var desc string
	var fields []*dg.MessageEmbedField

	if len(scores) == 0 {
		desc = "No one got any points."
	} else {
		desc = "__**Leaderboard**__\n"

		sort.Slice(scores, func(i, j int) bool {
			return scores[i].points > scores[j].points
		})

		var player, score strings.Builder
		for _, s := range scores {
			fmt.Fprintf(&player, "%s\n", s.player)
			fmt.Fprintf(&score, "%d\n", s.points)
		}

		fields = []*dg.MessageEmbedField{
			{
				Name:   "Player",
				Value:  player.String(),
				Inline: true,
			},
			{
				Name:   "Score",
				Value:  score.String(),
				Inline: true,
			},
		}
	}

	embed := &dg.MessageEmbed{
		Title:       "**Game Over!**",
		Description: desc,
		Fields:      fields,
		Footer: &dg.MessageEmbedFooter{
			Text: "Want to play Paintball? Enter: !pb",
		},
	}

	return embed
}
