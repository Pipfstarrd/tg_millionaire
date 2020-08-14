package main

import(
	"time"
	"fmt"
	"os"
	"io"
	"log"
	"encoding/csv"
	"bufio"
	"strconv"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	tb "gopkg.in/tucnak/telebot.v2"
)


type BotConf struct {
	m *tb.Message
	b *tb.Bot
	db *gorm.DB
	replyKeysMenu [][]tb.ReplyButton
	replyKeysGame [][]tb.ReplyButton
}


type Configuration struct {
	Key	string `gorm:"primary_key"`
	Value	string
}

func (Configuration) TableName() string {
	return "configuration";
}


type Question struct {
	Cost	uint64
	Number	uint64
	Question string
	Right	string
	Wrong1	string
	Wrong2	string
	Wrong3	string
	Media	string
}

type Game_statistics struct {
	Cost	uint
	Number	uint
	PlayerID uint64
	Answer	uint
	Time	uint64
}


// states for the State member of the Session struct
const (
	IDLE = iota
	LOADQUESTIONS
	LOAGQUESTIONSCONF
	GAME
)

type Session struct {
	UserID	int `gorm:"primary_key"`
	ChatID	int64
	IsAdmin	bool
	State	uint
	Timeout	uint
}

type GameSession struct {
	UserID	int `gorm:"primary_key"`
	ChatID	int64
	Question_number	int
	HadCalled	bool
	HadAudiance	bool
	HadFifty	bool
}

func main() {
	var bot BotConf
	var err error

	bot.db, err = gorm.Open("sqlite3", "server.db")
	if err != nil {
		panic("Failed to connect database")
	}

	bot.db.LogMode(true)
	defer bot.db.Close()

	if result := bot.db.AutoMigrate(&Configuration{}); result.Error != nil {
		log.Fatal("Failed to migrate Configuration scheme")
	}

/*	if result := bot.db.AutoMigrate(&Question{}); result.Error != nil {
		log.Fatal("Failed to migrate Question scheme")
	}

	if result := bot.db.AutoMigrate(&Game_statistics{}) result.Error != nil {
		log.Fatal("Failed to migrate Game_statistics scheme")
	}

	if result := bot.db.AutoMigrate(&Session{}); result.Error != nil {
		log.Fatal("Failed to migrate Session scheme")
	}*/


	// Create the handler for the API
	bot.b, err = tb.NewBot(tb.Settings{
		Token: os.Getenv("TG_MILLBOT_TKN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		fmt.Println("Error")
		fmt.Println(err.Error())
		return
	}

	// Creating Handles for events
	// Administration section first

	// Main Keyboard definition
	gameStartBtn := tb.ReplyButton{Text:"🏁 Новая игра 🏁"}
	helpBtn := tb.ReplyButton{Text:"🌚 Помощь 🌚"}
	ratingBtn := tb.ReplyButton{Text:"💯 Рейтинг игроков 💯"}
	bot.replyKeysMenu = [][]tb.ReplyButton{
		[]tb.ReplyButton{gameStartBtn},
		[]tb.ReplyButton{helpBtn},
		[]tb.ReplyButton{ratingBtn},
	}

		const admin_help_string string = "/helpadmin — эта справка\n" +
	"/addadmin @username — добавить админа в список админов\n" +
	"/removeadmin @username — удалить админа из списка админов\n" +
	"/loadquestions [прикрeплённый файл с вопросами в формате .csv] — загрузить новые " +
	"вопросы в базу данных. ВНИМАНИЕ! Старые вопросы удалятся, рейтинг обнулится."

	//TODO: АХТУНГ ГЛОБАЛЬНАЯ ПЕРЕМЕННАЯ! НУЖНО ПЕРЕДЕЛАТЬ РАБОТУ С DB

	// /helpadmin — prints a help message for admin
	bot.b.Handle("/helpadmin", func (m *tb.Message) {
		bot.b.Send(m.Sender, admin_help_string)
	})

	// /loadquestions — command for question import from attached CSV
	bot.b.Handle ("/loadquestions", func (m *tb.Message) {
		var session Session
		reply := bot.db.Find(&session, m.Sender.ID)

		if reply.RecordNotFound() {
			bot.UserNotFound(m)
			return
		} else {
			bot.b.Send(m.Sender, "Ну давай, присылай вопросы в .csv")
			session.State = LOADQUESTIONS
			bot.db.Save(session)
		}
	})

	bot.b.Handle(tb.OnDocument, func (m *tb.Message) {
		var session Session
		reply := bot.db.Find(&session, m.Sender.ID)

		if reply.RecordNotFound() {
			bot.UserNotFound(m)
			return
		} else {
			if session.State == LOADQUESTIONS {
				bot.b.Download(&m.Document.File, "quest1.csv")

				file, err := os.Open("quest1.csv")
				if err != nil {
					log.Fatal(err)
				}

				r := csv.NewReader(bufio.NewReader(file))
				for {
					record, err := r.Read()
					if err == io.EOF {
						break
					}
					if err != nil {
						log.Fatal(err)
					}

					cost, err := strconv.ParseUint(record[0], 10, 0)
					var number uint64 = 1

					question := Question{cost, number, record[1],
					record[2], record[3], record[4], record[5], ""}

					bot.db.NewRecord(question)
					bot.db.Create(&question)
				}

				session.State = IDLE
				bot.db.Save(session)
			} else {
				bot.b.Send(m.Sender, "Я от тебя не ждал такого, очень приятно :)")
			}
		}
	})

	// And here go user-available commands and game logic

	// Indefinite handler for all the garbage
	bot.b.Handle(tb.OnText, func(m *tb.Message) {
		var session Session
		reply := bot.db.Find(&session, m.Sender.ID)

		if reply.RecordNotFound() {
			bot.UserNotFound(m)
		} else {
			bot.b.Send(m.Sender, "Не дурачься (: . Список команд: кнопка \"Помощь\"," +
			"сбросить состояние бота: /start")
		}
	})

	const help_line string = "/start - инициализировать или сбросить бота\n" +
	"🏁 Новая игра 🏁 — начать новую игру\n" +
	"🌚 Помощь 🌚 — ещё раз прочитать эту справку\n" +
	"💯 Рейтинг игроков 💯 — посмотреть насколько ты близок ко дну\n"


	bot.b.Handle(&helpBtn, func (m *tb.Message) {
		bot.b.Send(m.Sender, help_line)
	})

	/*

type Session struct {
	UserID	int `gorm:"primary_key"`
	ChatID	int64
	IsAdmin	bool
	State	uint
	Timeout	uint
}
	*/
	bot.b.Handle(&gameStartBtn, func (m *tb.Message) {
		go bot.Game(m)
	})

	bot.b.Handle(&ratingBtn, func (m *tb.Message) {
		bot.b.Send(m.Sender, "Ещё не реализовано")
	})


	bot.b.Handle("/start", func (m *tb.Message) {
		bot.Reset(m)
	})

	bot.b.Start()
}


func (bot BotConf) UserNotFound(m *tb.Message) {
	bot.b.Send(m.Sender, "Тебя нет в списке игроков. Нажми " +
			"/start для инициализации.")
}

func (bot BotConf) Reset(m *tb.Message) {
	var session Session
	reply := bot.db.Find(&session, m.Sender.ID)
	if reply.RecordNotFound() {
		newSession := Session { m.Sender.ID, m.Chat.ID, false, IDLE, 120 }

		bot.db.NewRecord(newSession)
		bot.db.Create(&newSession)

		bot.b.Send(m.Sender, "Привет!", &tb.ReplyMarkup{
			ReplyKeyboard: bot.replyKeysMenu,
			ResizeReplyKeyboard: true,
		})
	} else {
		session.State = IDLE
		bot.db.Save(session)

		bot.b.Send(m.Sender, "Вы уже есть в списке игроков, для помощи наберите /help " +
		"или нажмите на кнопку \"Помощь\"", &tb.ReplyMarkup{
			ReplyKeyboard: bot.replyKeysMenu,
			ResizeReplyKeyboard: true,
		})
	}
}

func (bot BotConf) Game(m *tb.Message) {
	var session Session
	reply := bot.db.Find(&session, m.Sender.ID)
	if reply.RecordNotFound() {
		bot.UserNotFound(m)
	}

	session.State = GAME

	bot.db.AutoMigrate(&GameSession{})

	newgame := GameSession{m.Sender.ID, m.Chat.ID, 0, false, false, false}

	result := bot.db.NewRecord(newgame)
	fmt.Println(result)

	winBtn := tb.ReplyButton{Text: "WIN"}
	bot.replyKeysGame = [][]tb.ReplyButton  {
		[]tb.ReplyButton{winBtn},
	}

	bot.b.Send(m.Sender, "ИГРА", &tb.ReplyMarkup{
		ReplyKeyboard: bot.replyKeysGame,
		ResizeReplyKeyboard: true,
	})

/* type GameSession struct {
	UserID	int `gorm:"primary_key"`
	ChatID	int64
	Question_number	int
	HadCalled	bool
	HadAudiance	bool
	HadFifty	bool
	*/


	bot.b.Handle(&winBtn, func (m *tb.Message) {
	//	while session.State != 
		bot.b.Send(m.Sender, "Конец игры", &tb.ReplyMarkup{
		ReplyKeyboard: bot.replyKeysMenu,
		ResizeReplyKeyboard: true,
		})
	})
	

}
