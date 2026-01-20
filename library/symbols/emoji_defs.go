package symbols

// buildEmojiModule creates the emoji symbol module with common emoji characters.
// This is a subset of the full Typst codex emoji module, including the most
// commonly used emoji symbols.
func buildEmojiModule() *Module {
	m := &Module{
		Name:       "emoji",
		Symbols:    make(map[string]*Symbol),
		Submodules: make(map[string]*Module),
	}

	// Objects
	m.Symbols["abacus"] = singleSymbol("abacus", "🧮")
	m.Symbols["abc"] = singleSymbol("abc", "🔤")
	m.Symbols["abcd"] = singleSymbol("abcd", "🔡")
	m.Symbols["ABCD"] = singleSymbol("ABCD", "🔠")
	m.Symbols["accordion"] = singleSymbol("accordion", "🪗")
	m.Symbols["aesculapius"] = singleSymbol("aesculapius", "⚕")
	m.Symbols["airplane"] = newSymbol("airplane", map[string]string{
		"":        "✈",
		"landing": "🛬",
		"small":   "🛩",
		"takeoff": "🛫",
	})
	m.Symbols["alembic"] = singleSymbol("alembic", "⚗")
	m.Symbols["alien"] = newSymbol("alien", map[string]string{
		"":        "👽",
		"monster": "👾",
	})
	m.Symbols["ambulance"] = singleSymbol("ambulance", "🚑")
	m.Symbols["amphora"] = singleSymbol("amphora", "🏺")
	m.Symbols["anchor"] = singleSymbol("anchor", "⚓")
	m.Symbols["anger"] = singleSymbol("anger", "💢")
	m.Symbols["ant"] = singleSymbol("ant", "🐜")

	m.Symbols["apple"] = newSymbol("apple", map[string]string{
		"green": "🍏",
		"red":   "🍎",
	})

	m.Symbols["arm"] = newSymbol("arm", map[string]string{
		"mech":   "🦾",
		"muscle": "💪",
		"selfie": "🤳",
	})

	// Arrows (emoji version)
	m.Symbols["arrow"] = newSymbol("arrow", map[string]string{
		"r.filled":  "➡",
		"r.hook":    "↪",
		"r.soon":    "🔜",
		"l.filled":  "⬅",
		"l.hook":    "↩",
		"l.back":    "🔙",
		"l.end":     "🔚",
		"t.filled":  "⬆",
		"t.curve":   "⤴",
		"t.top":     "🔝",
		"b.filled":  "⬇",
		"b.curve":   "⤵",
		"l.r":       "↔",
		"l.r.on":    "🔛",
		"t.b":       "↕",
		"bl":        "↙",
		"br":        "↘",
		"tl":        "↖",
		"tr":        "↗",
	})

	m.Symbols["arrows"] = newSymbol("arrows", map[string]string{
		"cycle": "🔄",
	})

	m.Symbols["ast"] = newSymbol("ast", map[string]string{
		"":    "*",
		"box": "✳",
	})

	m.Symbols["atm"] = singleSymbol("atm", "🏧")
	m.Symbols["atom"] = singleSymbol("atom", "⚛")
	m.Symbols["aubergine"] = singleSymbol("aubergine", "🍆")
	m.Symbols["avocado"] = singleSymbol("avocado", "🥑")
	m.Symbols["axe"] = singleSymbol("axe", "🪓")

	// Baby
	m.Symbols["baby"] = newSymbol("baby", map[string]string{
		"":      "👶",
		"angel": "👼",
		"box":   "🚼",
	})

	m.Symbols["babybottle"] = singleSymbol("babybottle", "🍼")
	m.Symbols["backpack"] = singleSymbol("backpack", "🎒")
	m.Symbols["bacon"] = singleSymbol("bacon", "🥓")
	m.Symbols["badger"] = singleSymbol("badger", "🦡")
	m.Symbols["badminton"] = singleSymbol("badminton", "🏸")
	m.Symbols["bagel"] = singleSymbol("bagel", "🥯")
	m.Symbols["balloon"] = singleSymbol("balloon", "🎈")
	m.Symbols["banana"] = singleSymbol("banana", "🍌")
	m.Symbols["banjo"] = singleSymbol("banjo", "🪕")
	m.Symbols["bank"] = singleSymbol("bank", "🏦")
	m.Symbols["baseball"] = singleSymbol("baseball", "⚾")
	m.Symbols["basketball"] = newSymbol("basketball", map[string]string{
		"":     "⛹",
		"ball": "🏀",
	})
	m.Symbols["bat"] = singleSymbol("bat", "🦇")
	m.Symbols["bathtub"] = newSymbol("bathtub", map[string]string{
		"":     "🛀",
		"foam": "🛁",
	})

	m.Symbols["battery"] = newSymbol("battery", map[string]string{
		"":    "🔋",
		"low": "🪫",
	})

	m.Symbols["beach"] = newSymbol("beach", map[string]string{
		"palm":     "🏝",
		"umbrella": "🏖",
	})

	m.Symbols["bear"] = singleSymbol("bear", "🐻")
	m.Symbols["beaver"] = singleSymbol("beaver", "🦫")
	m.Symbols["bed"] = newSymbol("bed", map[string]string{
		"":       "🛏",
		"person": "🛌",
	})

	m.Symbols["bee"] = singleSymbol("bee", "🐝")
	m.Symbols["beer"] = newSymbol("beer", map[string]string{
		"":      "🍺",
		"clink": "🍻",
	})

	m.Symbols["bell"] = newSymbol("bell", map[string]string{
		"":     "🔔",
		"ding": "🛎",
	})

	// Common emoji faces
	m.Symbols["face"] = newSymbol("face", map[string]string{
		"smile":   "😊",
		"grin":    "😀",
		"laugh":   "😂",
		"wink":    "😉",
		"sad":     "😢",
		"cry":     "😭",
		"angry":   "😠",
		"think":   "🤔",
		"heart":   "😍",
		"cool":    "😎",
		"shock":   "😱",
		"sleep":   "😴",
		"sick":    "🤢",
		"nerd":    "🤓",
		"clown":   "🤡",
		"devil":   "😈",
		"ghost":   "👻",
		"skull":   "💀",
		"robot":   "🤖",
	})

	// Heart
	m.Symbols["heart"] = newSymbol("heart", map[string]string{
		"":        "❤",
		"orange":  "🧡",
		"yellow":  "💛",
		"green":   "💚",
		"blue":    "💙",
		"purple":  "💜",
		"black":   "🖤",
		"white":   "🤍",
		"brown":   "🤎",
		"broken":  "💔",
		"excl":    "❣",
		"sparkle": "💖",
		"grow":    "💗",
		"beat":    "💓",
		"two":     "💕",
		"arrow":   "💘",
		"ribbon":  "💝",
		"box":     "💟",
	})

	// Hands
	m.Symbols["hand"] = newSymbol("hand", map[string]string{
		"point.r": "👉",
		"point.l": "👈",
		"point.t": "👆",
		"point.b": "👇",
		"wave":    "👋",
		"ok":      "👌",
		"peace":   "✌",
		"rock":    "🤘",
		"thumb.t": "👍",
		"thumb.b": "👎",
		"fist":    "✊",
		"palm":    "🤚",
		"clap":    "👏",
		"fold":    "🙏",
		"shake":   "🤝",
		"write":   "✍",
	})

	// Weather
	m.Symbols["sun"] = newSymbol("sun", map[string]string{
		"":       "☀",
		"cloud":  "⛅",
		"behind": "🌤",
	})

	m.Symbols["moon"] = newSymbol("moon", map[string]string{
		"":        "🌙",
		"new":     "🌑",
		"waxing":  "🌒",
		"first":   "🌓",
		"gibbous": "🌔",
		"full":    "🌕",
		"waning":  "🌖",
		"last":    "🌗",
		"cresc":   "🌘",
		"face":    "🌛",
	})

	m.Symbols["cloud"] = newSymbol("cloud", map[string]string{
		"":        "☁",
		"rain":    "🌧",
		"snow":    "🌨",
		"storm":   "⛈",
		"light":   "🌩",
		"tornado": "🌪",
		"fog":     "🌫",
	})

	m.Symbols["rainbow"] = singleSymbol("rainbow", "🌈")
	m.Symbols["star"] = newSymbol("star", map[string]string{
		"":       "⭐",
		"glow":   "🌟",
		"shoot":  "🌠",
	})

	// Animals
	m.Symbols["dog"] = newSymbol("dog", map[string]string{
		"":     "🐕",
		"face": "🐶",
	})

	m.Symbols["cat"] = newSymbol("cat", map[string]string{
		"":     "🐈",
		"face": "🐱",
	})

	m.Symbols["bird"] = singleSymbol("bird", "🐦")
	m.Symbols["fish"] = singleSymbol("fish", "🐟")
	m.Symbols["snake"] = singleSymbol("snake", "🐍")
	m.Symbols["turtle"] = singleSymbol("turtle", "🐢")
	m.Symbols["frog"] = singleSymbol("frog", "🐸")
	m.Symbols["monkey"] = newSymbol("monkey", map[string]string{
		"":     "🐒",
		"face": "🐵",
	})

	m.Symbols["elephant"] = singleSymbol("elephant", "🐘")
	m.Symbols["lion"] = singleSymbol("lion", "🦁")
	m.Symbols["tiger"] = newSymbol("tiger", map[string]string{
		"":     "🐅",
		"face": "🐯",
	})

	m.Symbols["horse"] = newSymbol("horse", map[string]string{
		"":     "🐎",
		"face": "🐴",
	})

	m.Symbols["cow"] = newSymbol("cow", map[string]string{
		"":     "🐄",
		"face": "🐮",
	})

	m.Symbols["pig"] = newSymbol("pig", map[string]string{
		"":     "🐖",
		"face": "🐷",
	})

	m.Symbols["chicken"] = singleSymbol("chicken", "🐔")
	m.Symbols["rabbit"] = newSymbol("rabbit", map[string]string{
		"":     "🐇",
		"face": "🐰",
	})

	m.Symbols["wolf"] = singleSymbol("wolf", "🐺")
	m.Symbols["fox"] = singleSymbol("fox", "🦊")
	m.Symbols["unicorn"] = singleSymbol("unicorn", "🦄")
	m.Symbols["dragon"] = newSymbol("dragon", map[string]string{
		"":     "🐉",
		"face": "🐲",
	})

	// Food
	m.Symbols["pizza"] = singleSymbol("pizza", "🍕")
	m.Symbols["burger"] = singleSymbol("burger", "🍔")
	m.Symbols["fries"] = singleSymbol("fries", "🍟")
	m.Symbols["hotdog"] = singleSymbol("hotdog", "🌭")
	m.Symbols["taco"] = singleSymbol("taco", "🌮")
	m.Symbols["burrito"] = singleSymbol("burrito", "🌯")
	m.Symbols["sandwich"] = singleSymbol("sandwich", "🥪")
	m.Symbols["sushi"] = singleSymbol("sushi", "🍣")
	m.Symbols["ramen"] = singleSymbol("ramen", "🍜")
	m.Symbols["cake"] = newSymbol("cake", map[string]string{
		"":         "🎂",
		"birthday": "🎂",
		"slice":    "🍰",
		"short":    "🍰",
	})

	m.Symbols["cookie"] = singleSymbol("cookie", "🍪")
	m.Symbols["donut"] = singleSymbol("donut", "🍩")
	m.Symbols["icecream"] = newSymbol("icecream", map[string]string{
		"":     "🍨",
		"soft": "🍦",
	})

	m.Symbols["coffee"] = singleSymbol("coffee", "☕")
	m.Symbols["tea"] = singleSymbol("tea", "🍵")

	// Sports
	m.Symbols["soccer"] = singleSymbol("soccer", "⚽")
	m.Symbols["football"] = singleSymbol("football", "🏈")
	m.Symbols["tennis"] = singleSymbol("tennis", "🎾")
	m.Symbols["golf"] = singleSymbol("golf", "⛳")
	m.Symbols["trophy"] = singleSymbol("trophy", "🏆")
	m.Symbols["medal"] = newSymbol("medal", map[string]string{
		"gold":   "🥇",
		"silver": "🥈",
		"bronze": "🥉",
		"sports": "🏅",
	})

	// Celebration
	m.Symbols["party"] = singleSymbol("party", "🎉")
	m.Symbols["confetti"] = singleSymbol("confetti", "🎊")
	m.Symbols["fireworks"] = singleSymbol("fireworks", "🎆")
	m.Symbols["sparkler"] = singleSymbol("sparkler", "🎇")
	m.Symbols["gift"] = singleSymbol("gift", "🎁")

	// Tools and Objects
	m.Symbols["hammer"] = newSymbol("hammer", map[string]string{
		"":       "🔨",
		"wrench": "🛠",
	})

	m.Symbols["wrench"] = singleSymbol("wrench", "🔧")
	m.Symbols["gear"] = singleSymbol("gear", "⚙")
	m.Symbols["key"] = newSymbol("key", map[string]string{
		"":    "🔑",
		"old": "🗝",
	})

	m.Symbols["lock"] = newSymbol("lock", map[string]string{
		"":     "🔒",
		"open": "🔓",
	})

	m.Symbols["bulb"] = singleSymbol("bulb", "💡")
	m.Symbols["magnify"] = newSymbol("magnify", map[string]string{
		"l": "🔍",
		"r": "🔎",
	})

	m.Symbols["book"] = newSymbol("book", map[string]string{
		"":       "📖",
		"closed": "📕",
		"green":  "📗",
		"blue":   "📘",
		"orange": "📙",
		"stack":  "📚",
	})

	m.Symbols["computer"] = newSymbol("computer", map[string]string{
		"":        "💻",
		"desktop": "🖥",
	})

	m.Symbols["phone"] = newSymbol("phone", map[string]string{
		"":       "📱",
		"old":    "☎",
		"off":    "📴",
		"vibra":  "📳",
	})

	m.Symbols["camera"] = newSymbol("camera", map[string]string{
		"":     "📷",
		"film": "🎥",
	})

	m.Symbols["mail"] = newSymbol("mail", map[string]string{
		"":        "📧",
		"inbox":   "📥",
		"outbox":  "📤",
		"open":    "📬",
		"closed":  "📪",
		"love":    "💌",
	})

	m.Symbols["folder"] = newSymbol("folder", map[string]string{
		"":     "📁",
		"open": "📂",
	})

	// Symbols
	m.Symbols["check"] = newSymbol("check", map[string]string{
		"":      "✔",
		"white": "✅",
	})

	m.Symbols["cross"] = newSymbol("cross", map[string]string{
		"":    "❌",
		"red": "❌",
	})

	m.Symbols["question"] = newSymbol("question", map[string]string{
		"":    "❓",
		"red": "❓",
	})

	m.Symbols["exclaim"] = newSymbol("exclaim", map[string]string{
		"":    "❗",
		"red": "❗",
	})

	m.Symbols["warning"] = singleSymbol("warning", "⚠")
	m.Symbols["stop"] = singleSymbol("stop", "🛑")
	m.Symbols["recycle"] = singleSymbol("recycle", "♻")

	// Flags (common ones)
	m.Symbols["flag"] = newSymbol("flag", map[string]string{
		"chequered": "🏁",
		"crossed":   "🎌",
		"black":     "🏴",
		"white":     "🏳",
		"rainbow":   "🏳️‍🌈",
		"pirate":    "🏴‍☠️",
	})

	// Misc
	m.Symbols["fire"] = singleSymbol("fire", "🔥")
	m.Symbols["water"] = singleSymbol("water", "💧")
	m.Symbols["wind"] = singleSymbol("wind", "💨")
	m.Symbols["snowflake"] = singleSymbol("snowflake", "❄")
	m.Symbols["lightning"] = singleSymbol("lightning", "⚡")
	m.Symbols["boom"] = singleSymbol("boom", "💥")
	m.Symbols["sparkle"] = singleSymbol("sparkle", "✨")
	m.Symbols["rocket"] = singleSymbol("rocket", "🚀")
	m.Symbols["car"] = singleSymbol("car", "🚗")
	m.Symbols["bus"] = singleSymbol("bus", "🚌")
	m.Symbols["train"] = singleSymbol("train", "🚃")
	m.Symbols["ship"] = singleSymbol("ship", "🚢")
	m.Symbols["bike"] = singleSymbol("bike", "🚲")
	m.Symbols["house"] = singleSymbol("house", "🏠")
	m.Symbols["tree"] = newSymbol("tree", map[string]string{
		"":          "🌳",
		"evergreen": "🌲",
		"palm":      "🌴",
	})

	m.Symbols["flower"] = newSymbol("flower", map[string]string{
		"":         "🌸",
		"cherry":   "🌸",
		"rose":     "🌹",
		"tulip":    "🌷",
		"sunflower":"🌻",
		"bouquet":  "💐",
	})

	m.Symbols["clock"] = newSymbol("clock", map[string]string{
		"":    "🕐",
		"1":   "🕐",
		"2":   "🕑",
		"3":   "🕒",
		"4":   "🕓",
		"5":   "🕔",
		"6":   "🕕",
		"7":   "🕖",
		"8":   "🕗",
		"9":   "🕘",
		"10":  "🕙",
		"11":  "🕚",
		"12":  "🕛",
	})

	m.Symbols["100"] = singleSymbol("100", "💯")
	m.Symbols["zzz"] = singleSymbol("zzz", "💤")

	return m
}
