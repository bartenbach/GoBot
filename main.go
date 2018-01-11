package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/proxa/GoBot/markov"

	"github.com/whyrusleeping/hellabot"
	log "gopkg.in/inconshreveable/log15.v2"
)

var botName = "golangbot"
var serv = flag.String("server", "chat.freenode.net:6667", "hostname and port for irc server to connect to")
var nick = flag.String("nick", botName, "nickname for the bot")

func main() {
	flag.Parse()
	createTable()

	hijackSession := func(bot *hbot.Bot) {
		bot.HijackSession = true
	}
	channels := func(bot *hbot.Bot) {
		bot.Channels = []string{"#afterlunch"}
	}
	irc, err := hbot.NewBot(*serv, *nick, hijackSession, channels)
	if err != nil {
		panic(err)
	}

	irc.AddTrigger(SayInfoMessage)
	irc.AddTrigger(LogMessage)
	irc.AddTrigger(MarkovChain)
	irc.Logger.SetHandler(log.StdoutHandler)

	irc.Run()
	fmt.Println("Bot shutting down.")
}

// LogMessage logs all messages from chat to the database for chaining later.
var LogMessage = hbot.Trigger{
	func(bot *hbot.Bot, m *hbot.Message) bool {
		return !strings.Contains(m.From, ".") && m.From != botName && m.From != "" && !strings.HasPrefix(m.Content, "-")
	},
	func(irc *hbot.Bot, m *hbot.Message) bool {
		writeMessageToDatabase(m.Content)
		return true
	},
}

// SayInfoMessage is a test function.
var SayInfoMessage = hbot.Trigger{
	func(bot *hbot.Bot, m *hbot.Message) bool {
		return m.Command == "PRIVMSG" && m.Content == "-test"
	},
	func(irc *hbot.Bot, m *hbot.Message) bool {
		irc.Reply(m, "Tyler is a big faggot")
		return false
	},
}

// MarkovChain is the on-demand way to start the markov chain.
var MarkovChain = hbot.Trigger{
	func(bot *hbot.Bot, m *hbot.Message) bool {
		return m.Command == "PRIVMSG" && m.Content == "-mkv"
	},
	func(irc *hbot.Bot, m *hbot.Message) bool {
		reply := getMarkovText()
		irc.Reply(m, reply)
		return false
	},
}

func getMarkovText() string {
	data := getMessageFromDatabase()
	fmt.Println("DATA: ", data)
	result := markov.DoMarkovChain(data)
	fmt.Println("RESULT: ", result)
	return result
}

func createTable() {
	// connect to sql database named 'gobot'
	db, err := sql.Open("mysql", "gobot:test@/gobot?charset=utf8")
	checkErr(err)

	// create the table
	createTable := string("CREATE TABLE IF NOT EXISTS `messages` (`message` VARCHAR(450) NOT NULL);")
	stmt, err := db.Prepare(createTable)
	checkErr(err)
	res, err := stmt.Exec()
	checkErr(err)
	fmt.Println(res)
}

func writeMessageToDatabase(msg string) {
	db, err := sql.Open("mysql", "gobot:test@/gobot?charset=utf8")
	checkErr(err)

	stmt, err := db.Prepare("INSERT messages SET message=?")
	checkErr(err)

	res, err := stmt.Exec(msg)
	fmt.Println(res)
	checkErr(err)
}

func getMessageFromDatabase() string {
	db, err := sql.Open("mysql", "gobot:test@/gobot?charset=utf8")
	checkErr(err)
	rows, err := db.Query("SELECT * FROM messages ORDER BY RAND()")
	if err != nil {
		panic(err)
	}
	var tmp string
	defer rows.Close()
	var buffer bytes.Buffer
	for rows.Next() {
		err := rows.Scan(&tmp)
		buffer.WriteString(tmp + " ")
		if err != nil {
			fmt.Println(err)
		}
	}
	err = rows.Err()
	if err != nil {
		fmt.Println(err)
	}
	return buffer.String()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
