package main

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/natefinch/lumberjack"
)

var (
	host     = os.Getenv("RAIDBOT_HOST")
	port     = os.Getenv("RAIDBOT_PORT")
	user     = os.Getenv("RAIDBOT_USER")
	password = os.Getenv("RAIDBOT_PASS")
	botKey   = os.Getenv("RAIDBOT_TOKEN")
	logDir   = os.Getenv("RAIDBOT_LOGDIR") //with / at the end
	dbname   = "pnzraid"
)

var db *sqlx.DB
var bot *tgbotapi.BotAPI
var sender BotSender
var sendUpdates = make(map[int64]bool)

func main() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   logDir + "remoteraidbot.log",
		MaxSize:    5, // megabytes
		MaxBackups: 3,
		MaxAge:     7,    //days
		Compress:   true, // disabled by default
	})

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	var err error
	db, err = sqlx.Open("postgres", psqlInfo)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	var needInit bool
	needInit = !initializedDB()

	if needInit {
		fmt.Println("need init")
		initDB()
		needInit = !initializedDB()

		if needInit {
			log.Panic("db not initialized")
		} else {
			fmt.Println("init ok")
		}
	} else {
		fmt.Println("init ok")
	}

	/*
		err = db.Ping()
		if err != nil {
			panic(err)
		}
	*/
	bot, err = tgbotapi.NewBotAPI(botKey)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	//bot.RemoveWebhook()

	updates, err := bot.GetUpdatesChan(u)

	if err != nil {
		panic("bot.GetUpdatesChan()")
	}

	sender.Init(bot)

	go updateUserInfo()

	for update := range updates {
		if update.CallbackQuery != nil && nil != update.CallbackQuery.From {

			userID := User(update.CallbackQuery.From.ID)
			var chatID *int64
			var msgID *int
			if nil != update.CallbackQuery.Message {
				if nil != update.CallbackQuery.Message.Chat {
					chatID = &update.CallbackQuery.Message.Chat.ID
				}
				msgID = &update.CallbackQuery.Message.MessageID
			}

			if !userID.IsRegistered() {
				registerUser(userID, update.CallbackQuery.From)
			}

			bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Обработка..."))
			processCommand(userID, chatID, msgID, update.CallbackQuery.InlineMessageID, update.CallbackQuery.Data)

			if chatID != nil && msgID != nil {
				log.Printf("%d:%d [%s]~ %s\r\n", *chatID, *msgID, update.CallbackQuery.From.UserName, update.CallbackQuery.Data)
			} else {
				log.Printf("[%s]~ %s\r\n", update.CallbackQuery.From.UserName, update.CallbackQuery.Data)
			}
		}

		if update.Message != nil && update.Message.From != nil {
			userID := User(update.Message.From.ID)
			var chatID *int64
			var msgID *int

			if nil != update.Message.Chat {
				chatID = &update.Message.Chat.ID
			}
			msgID = &update.Message.MessageID

			if !userID.IsRegistered() {
				registerUser(userID, update.Message.From)
			}

			processCommand(userID, chatID, nil, "", update.Message.Text)
			if chatID != nil && msgID != nil {
				log.Printf("%d:%d [%s] %s\r\n", *chatID, *msgID, update.Message.From.UserName, update.Message.Text)
			} else {
				log.Printf("[%s] %s\r\n", update.Message.From.UserName, update.Message.Text)
			}
		}
		if update.InlineQuery != nil {
			processInlineQuery(update.InlineQuery.ID, update.InlineQuery.Query)
		}
	}
}

func registerUser(user User, tgu *tgbotapi.User) {
	var name string
	if nil == tgu {
		name = fmt.Sprintf("user%d", user)
	} else {
		if len(tgu.UserName) > 0 {
			name = tgu.UserName
		} else {
			if (len(tgu.FirstName) > 0) || (len(tgu.LastName) > 0) {
				name = tgu.FirstName + " " + tgu.LastName
			} else {
				name = fmt.Sprintf("user%d", user)
			}
		}
	}
	fmt.Printf("New user %d known as %s\r\n", user, name)
	user.SetName(name)
}

func tableExists(tableName string) (result bool) {
	var res []bool
	err := db.Select(&res, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1);", tableName)
	if err != nil {
		log.Panic(err)
	}
	if len(res) > 0 {
		result = res[0]
	}
	return
}

func seqExists(seqName string) (result bool) {
	var res []bool
	err := db.Select(&res, "SELECT EXISTS (SELECT 0 FROM pg_class where relname = $1);", seqName)
	if err != nil {
		log.Panic(err)
	}
	if len(res) > 0 {
		result = res[0]
	}
	return
}

func initializedDB() (result bool) {
	allTables := []string{"players", "chats", "raids", "votes", "groupmessages"}

	for _, str := range allTables {
		result = tableExists(str)
		if !result {
			fmt.Println("table " + str + " not exists")
			break
		}
	}

	allSeqs := []string{"raid_id_seq"}
	for _, str := range allSeqs {
		result = seqExists(str)

		if !result {
			fmt.Println("sequence " + str + " not exists")
			break
		}
	}

	return
}

func initExec(str string) {
	_, err := db.Exec(str)
	if err != nil {
		log.Panic(err)
	}
}

func initDB() {

	if !seqExists("raid_id_seq") {
		initExec(`
			CREATE SEQUENCE public.raid_id_seq
			INCREMENT 1
			START 1
			MINVALUE 1
			MAXVALUE 9223372036854775807
			CACHE 1;`)

		initExec(`ALTER SEQUENCE public.raid_id_seq OWNER TO ` + dbname + `;`)
	}

	if !tableExists("chats") {
		initExec(`CREATE TABLE public.chats
		(
			raid_id integer NOT NULL,
			chat_id bigint NOT NULL,
			msg_id integer NOT NULL
		)`)
	}

	if !tableExists("groupmessages") {
		initExec(`CREATE TABLE public.groupmessages
	(
		raid_id integer NOT NULL,
		inline_msg_id text COLLATE pg_catalog."default" NOT NULL
	)`)
	}

	if !tableExists("players") {
		initExec(`CREATE TABLE public.players
		(
			user_id integer NOT NULL,
			pogoname text COLLATE pg_catalog."default" NOT NULL,
			pogocode text COLLATE pg_catalog."default",
			notif boolean NOT NULL DEFAULT false,
			CONSTRAINT users_pkey PRIMARY KEY (user_id)
		)`)
	}

	if !tableExists("raids") {
		initExec(`CREATE TABLE public.raids
		(
			raid_id integer NOT NULL DEFAULT nextval('raid_id_seq'::regclass),
			raid_info text COLLATE pg_catalog."default" NOT NULL,
			chat_id bigint NOT NULL,
			msg_id integer,
			user_id integer NOT NULL,
			finished boolean NOT NULL,
			CONSTRAINT raids_pkey PRIMARY KEY (raid_id)
		)`)
	}

	if !tableExists("votes") {
		initExec(`CREATE TABLE public.votes
		(
			raid_id integer NOT NULL,
			raid_role text COLLATE pg_catalog."default" NOT NULL,
			user_id integer NOT NULL
		)`)
	}

	//initExec(`TABLESPACE pg_default;`)

	initExec(`ALTER TABLE public.chats	OWNER to ` + dbname + `;`)
	initExec(`ALTER TABLE public.groupmessages	OWNER to ` + dbname + `;`)
	initExec(`ALTER TABLE public.players	OWNER to ` + dbname + `;`)
	initExec(`ALTER TABLE public.raids	OWNER to ` + dbname + `;`)
	initExec(`ALTER TABLE public.votes	OWNER to ` + dbname + `;`)
	return
}
