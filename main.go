package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
)

func main() {
	openaitk := os.Getenv("OPENAI_TOKEN")

	// Bot の作成
	discordtk := os.Getenv("DISCORD_BOT_TOKEN")
	// 初期プロンプトを自由に設定できるようにはしている
	initprompt := ""
	// チャットモデルの設定
	model := os.Getenv("OPENAI_MODEL")
	// 投稿するチャンネルID
	cid := os.Getenv("DISCORD_CHANNEL_ID")
	// API のエンドポイントURL
	baseUrl := os.Getenv("OPENAI_BASEURL")
	// 設定されていなければ openai のURLを使用する
	if baseUrl == "" {
		baseUrl = "https://api.openai.com/v1/"
	}
	log.Printf("Info: API baseurl is %s", baseUrl)
	log.Println("Info: Bot starting...")
	log.Printf("Info: OpenAI model is %s", model)

	// 記憶の保持期間
	days, err := strconv.Atoi(os.Getenv("HISTORY_DAYS"))
	if err != nil {
		log.Fatal("Error: HISTORY_DAYS is invalid, value: ", days)
	}
	// history.json への保存間隔(デフォルト1分)
	saveRotateMin, err := strconv.Atoi(os.Getenv("SAVE_ROTATE_MIN"))
	if err != nil {
		log.Fatal("Error: SAVE_ROTATE_MIN is invalid, value: ", saveRotateMin)
	}
	// history.json の保存ディレクトリ
	historyDir := os.Getenv("HISTORY_DIR")
	// ディレクトリがなければ作成する
	if _, err := os.Stat(historyDir); err != nil {
		log.Printf("Info: History dir not exist, creating...")
		os.Mkdir(historyDir, os.ModePerm)
	}
	if openaitk == "" || model == "" {
		log.Fatal("Error: OpenAI token or OpenAI model not set")
		return
	}

	if discordtk == "" {
		log.Fatal("Error: Discord token not set")
		return
	}
	hm := NewHistoryManager(historyDir, days, saveRotateMin)
	openaisv := NewOpenAiService(openaitk, baseUrl)
	bot := NewDiscordBot(discordtk, model, initprompt, hm)

	bot.Session.AddHandler(readyHandler(bot, openaisv, cid))
	bot.Session.AddHandler(messageCreateHandler(bot, cid, openaisv))
	bot.Session.AddHandler(forgetCommandHandler(bot))
	bot.Session.AddHandler(forgetConfirmHandler(bot))
	bot.Session.Open()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch
	hm.Stop()
	err = bot.Session.Close()

	log.Println("Info: Bot stopping...")
	if err != nil {
		log.Printf("Error: Something went wrong when session closing: %f", err)
	}

}
