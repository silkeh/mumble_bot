package telegram

import (
	"github.com/tucnak/telebot"
)

// Stickers contains a mapping of names to stickers
var Stickers = map[string]*telebot.Sticker{
	"welcome": {
		File: telebot.File{FileID: "CAADBAADdQADwWSEAAGLkxO7uQz7zwI"},
	},
	"groeten": {
		File: telebot.File{FileID: "CAADBAADdQADwWSEAAGLkxO7uQz7zwI"},
	},
	"help": {
		File: telebot.File{FileID: "CAADBAADYAADwWSEAAETwmKpJEG89wI"},
	},
	"gwyf": {
		File: telebot.File{FileID: "CAADBAADWwADwWSEAAHTDGwW7V7VuwI"},
	},
	"dingen": {
		File: telebot.File{FileID: "CAADBAADWQADwWSEAAF15hZh8XuaGQI"},
	},
	"bind": {
		File: telebot.File{FileID: "CAADBAADYgADwWSEAAEcj_frlPhASwI"},
	},
	"hmmm": {
		File: telebot.File{FileID: "CAADBAADZAADwWSEAAGLrm_RmieQPwI"},
	},
	"dat": {
		File: telebot.File{FileID: "CAADBAADZgADwWSEAAHUpk5Q7DzAhAI"},
	},
}
