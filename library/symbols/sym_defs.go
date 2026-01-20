package symbols

// buildSymModule creates the main symbol module with all general and math symbols.
// Symbols are organized by category following the Typst codex structure.
func buildSymModule() *Module {
	m := &Module{
		Name:       "sym",
		Symbols:    make(map[string]*Symbol),
		Submodules: make(map[string]*Module),
	}

	// Control characters
	m.Symbols["wj"] = singleSymbol("wj", "\u2060")
	m.Symbols["zwj"] = singleSymbol("zwj", "\u200D")
	m.Symbols["zwnj"] = singleSymbol("zwnj", "\u200C")
	m.Symbols["zws"] = singleSymbol("zws", "\u200B")
	m.Symbols["lrm"] = singleSymbol("lrm", "\u200E")
	m.Symbols["rlm"] = singleSymbol("rlm", "\u200F")

	// Spaces
	m.Symbols["space"] = newSymbol("space", map[string]string{
		"":               " ",
		"nobreak":        "\u00A0",
		"nobreak.narrow": "\u202F",
		"en":             "\u2002",
		"quad":           "\u2003",
		"third":          "\u2004",
		"quarter":        "\u2005",
		"sixth":          "\u2006",
		"med":            "\u205F",
		"fig":            "\u2007",
		"punct":          "\u2008",
		"thin":           "\u2009",
		"hair":           "\u200A",
	})

	// Delimiters
	m.Symbols["paren"] = newSymbol("paren", map[string]string{
		"l":         "(",
		"l.flat":    "⟮",
		"l.closed":  "⦇",
		"l.stroked": "⦅",
		"r":         ")",
		"r.flat":    "⟯",
		"r.closed":  "⦈",
		"r.stroked": "⦆",
		"t":         "⏜",
		"b":         "⏝",
	})

	m.Symbols["brace"] = newSymbol("brace", map[string]string{
		"l":         "{",
		"l.stroked": "⦃",
		"r":         "}",
		"r.stroked": "⦄",
		"t":         "⏞",
		"b":         "⏟",
	})

	m.Symbols["bracket"] = newSymbol("bracket", map[string]string{
		"l":         "[",
		"l.tick.t":  "⦍",
		"l.tick.b":  "⦏",
		"l.stroked": "⟦",
		"r":         "]",
		"r.tick.t":  "⦐",
		"r.tick.b":  "⦎",
		"r.stroked": "⟧",
		"t":         "⎴",
		"b":         "⎵",
	})

	m.Symbols["shell"] = newSymbol("shell", map[string]string{
		"l":         "❲",
		"l.stroked": "⟬",
		"l.filled":  "⦗",
		"r":         "❳",
		"r.stroked": "⟭",
		"r.filled":  "⦘",
		"t":         "⏠",
		"b":         "⏡",
	})

	m.Symbols["bag"] = newSymbol("bag", map[string]string{
		"l": "⟅",
		"r": "⟆",
	})

	m.Symbols["mustache"] = newSymbol("mustache", map[string]string{
		"l": "⎰",
		"r": "⎱",
	})

	m.Symbols["bar"] = newSymbol("bar", map[string]string{
		"v":        "|",
		"v.double": "‖",
		"v.triple": "⦀",
		"v.broken": "¦",
		"v.o":      "⦶",
		"h":        "―",
	})

	m.Symbols["fence"] = newSymbol("fence", map[string]string{
		"l":        "⧘",
		"l.double": "⧚",
		"r":        "⧙",
		"r.double": "⧛",
		"dotted":   "⦙",
	})

	m.Symbols["chevron"] = newSymbol("chevron", map[string]string{
		"l":        "⟨",
		"l.curly":  "⧼",
		"l.dot":    "⦑",
		"l.closed": "⦉",
		"l.double": "⟪",
		"r":        "⟩",
		"r.curly":  "⧽",
		"r.dot":    "⦒",
		"r.closed": "⦊",
		"r.double": "⟫",
	})

	m.Symbols["ceil"] = newSymbol("ceil", map[string]string{
		"l": "⌈",
		"r": "⌉",
	})

	m.Symbols["floor"] = newSymbol("floor", map[string]string{
		"l": "⌊",
		"r": "⌋",
	})

	m.Symbols["corner"] = newSymbol("corner", map[string]string{
		"l.t": "⌜",
		"l.b": "⌞",
		"r.t": "⌝",
		"r.b": "⌟",
	})

	// Punctuation
	m.Symbols["amp"] = newSymbol("amp", map[string]string{
		"":    "&",
		"inv": "⅋",
	})

	m.Symbols["ast"] = newSymbol("ast", map[string]string{
		"op":     "∗",
		"op.o":   "⊛",
		"basic":  "*",
		"low":    "⁎",
		"double": "⁑",
		"triple": "⁂",
		"square": "⧆",
	})

	m.Symbols["at"] = singleSymbol("at", "@")

	m.Symbols["backslash"] = newSymbol("backslash", map[string]string{
		"":    "\\",
		"o":   "⦸",
		"not": "⧷",
	})

	m.Symbols["co"] = singleSymbol("co", "℅")

	m.Symbols["colon"] = newSymbol("colon", map[string]string{
		"":          ":",
		"currency":  "₡",
		"double":    "∷",
		"tri":       "⁝",
		"tri.op":    "⫶",
		"eq":        "≔",
		"double.eq": "⩴",
	})

	m.Symbols["comma"] = newSymbol("comma", map[string]string{
		"":    ",",
		"inv": "⸲",
		"rev": "⹁",
	})

	m.Symbols["dagger"] = newSymbol("dagger", map[string]string{
		"":       "†",
		"double": "‡",
		"triple": "⹋",
		"l":      "⸶",
		"r":      "⸷",
		"inv":    "⸸",
	})

	m.Symbols["dash"] = newSymbol("dash", map[string]string{
		"en":          "–",
		"em":          "—",
		"em.two":      "⸺",
		"em.three":    "⸻",
		"fig":         "‒",
		"colon":       "∹",
		"o":           "⊝",
		"wave":        "〜",
		"wave.double": "〰",
	})

	m.Symbols["dot"] = newSymbol("dot", map[string]string{
		"op":     "⋅",
		"basic":  ".",
		"c":      "·",
		"o":      "⊙",
		"o.big":  "⨀",
		"square": "⊡",
		"double": "¨",
		"triple": "\u20DB",
		"quad":   "\u20DC",
	})

	m.Symbols["excl"] = newSymbol("excl", map[string]string{
		"":       "!",
		"double": "‼",
		"inv":    "¡",
		"quest":  "⁉",
	})

	m.Symbols["quest"] = newSymbol("quest", map[string]string{
		"":       "?",
		"double": "⁇",
		"excl":   "⁈",
		"inv":    "¿",
	})

	m.Symbols["interrobang"] = newSymbol("interrobang", map[string]string{
		"":    "‽",
		"inv": "⸘",
	})

	m.Symbols["hash"] = singleSymbol("hash", "#")

	m.Symbols["hyph"] = newSymbol("hyph", map[string]string{
		"":        "‐",
		"minus":   "-",
		"nobreak": "\u2011",
		"point":   "‧",
		"soft":    "\u00AD",
	})

	m.Symbols["numero"] = singleSymbol("numero", "№")
	m.Symbols["percent"] = singleSymbol("percent", "%")
	m.Symbols["permille"] = singleSymbol("permille", "‰")
	m.Symbols["permyriad"] = singleSymbol("permyriad", "‱")

	m.Symbols["pilcrow"] = newSymbol("pilcrow", map[string]string{
		"":    "¶",
		"rev": "⁋",
	})

	m.Symbols["section"] = singleSymbol("section", "§")

	m.Symbols["semi"] = newSymbol("semi", map[string]string{
		"":    ";",
		"inv": "⸵",
		"rev": "⁏",
	})

	m.Symbols["slash"] = newSymbol("slash", map[string]string{
		"":       "/",
		"o":      "⊘",
		"double": "⫽",
		"triple": "⫻",
		"big":    "⧸",
	})

	m.Symbols["dots"] = newSymbol("dots", map[string]string{
		"h.c":  "⋯",
		"h":    "…",
		"v":    "⋮",
		"down": "⋱",
		"up":   "⋰",
	})

	m.Symbols["tilde"] = newSymbol("tilde", map[string]string{
		"op":         "∼",
		"basic":      "~",
		"dot":        "⩪",
		"eq":         "≃",
		"eq.not":     "≄",
		"eq.rev":     "⋍",
		"equiv":      "≅",
		"equiv.not":  "≇",
		"nequiv":     "≆",
		"not":        "≁",
		"rev":        "∽",
		"rev.equiv":  "≌",
		"triple":     "≋",
	})

	// Accents, quotes, and primes
	m.Symbols["acute"] = newSymbol("acute", map[string]string{
		"":       "´",
		"double": "˝",
	})

	m.Symbols["breve"] = singleSymbol("breve", "˘")
	m.Symbols["caret"] = singleSymbol("caret", "‸")
	m.Symbols["caron"] = singleSymbol("caron", "ˇ")
	m.Symbols["hat"] = singleSymbol("hat", "^")
	m.Symbols["diaer"] = singleSymbol("diaer", "¨")
	m.Symbols["grave"] = singleSymbol("grave", "`")
	m.Symbols["macron"] = singleSymbol("macron", "¯")

	m.Symbols["quote"] = newSymbol("quote", map[string]string{
		"double":           "\"",
		"single":           "'",
		"l.double":         "\u201C", // "
		"l.single":         "\u2018", // '
		"r.double":         "\u201D", // "
		"r.single":         "\u2019", // '
		"chevron.l.double": "«",
		"chevron.l.single": "‹",
		"chevron.r.double": "»",
		"chevron.r.single": "›",
		"high.double":      "\u201F", // ‟
		"high.single":      "\u201B", // ‛
		"low.double":       "\u201E", // „
		"low.single":       "\u201A", // ‚
	})

	m.Symbols["prime"] = newSymbol("prime", map[string]string{
		"":           "′",
		"rev":        "‵",
		"double":     "″",
		"double.rev": "‶",
		"triple":     "‴",
		"triple.rev": "‷",
		"quad":       "⁗",
	})

	// Arithmetic
	m.Symbols["plus"] = newSymbol("plus", map[string]string{
		"":         "+",
		"o":        "⊕",
		"o.l":      "⨭",
		"o.r":      "⨮",
		"o.arrow":  "⟴",
		"o.big":    "⨁",
		"dot":      "∔",
		"double":   "⧺",
		"minus":    "±",
		"square":   "⊞",
		"triangle": "⨹",
		"triple":   "⧻",
	})

	m.Symbols["minus"] = newSymbol("minus", map[string]string{
		"":         "−",
		"o":        "⊖",
		"dot":      "∸",
		"plus":     "∓",
		"square":   "⊟",
		"tilde":    "≂",
		"triangle": "⨺",
	})

	m.Symbols["div"] = newSymbol("div", map[string]string{
		"":          "÷",
		"o":         "⨸",
		"slanted.o": "⦼",
	})

	m.Symbols["times"] = newSymbol("times", map[string]string{
		"":         "×",
		"big":      "⨉",
		"o":        "⊗",
		"o.l":      "⨴",
		"o.r":      "⨵",
		"o.hat":    "⨶",
		"o.big":    "⨂",
		"div":      "⋇",
		"three.l":  "⋋",
		"three.r":  "⋌",
		"l":        "⋉",
		"r":        "⋊",
		"square":   "⊠",
		"triangle": "⨻",
	})

	m.Symbols["ratio"] = singleSymbol("ratio", "∶")

	// Relations
	m.Symbols["eq"] = newSymbol("eq", map[string]string{
		"":           "=",
		"star":       "≛",
		"o":          "⊜",
		"colon":      "≕",
		"dots":       "≑",
		"dots.down":  "≒",
		"dots.up":    "≓",
		"def":        "≝",
		"delta":      "≜",
		"equi":       "≚",
		"est":        "≙",
		"gt":         "⋝",
		"lt":         "⋜",
		"m":          "≞",
		"not":        "≠",
		"prec":       "⋞",
		"quest":      "≟",
		"succ":       "⋟",
		"triple":     "≡",
		"triple.not": "≢",
		"quad":       "≣",
	})

	m.Symbols["gt"] = newSymbol("gt", map[string]string{
		"":              ">",
		"o":             "⧁",
		"dot":           "⋗",
		"approx":        "⪆",
		"arc":           "⪧",
		"arc.eq":        "⪩",
		"double":        "≫",
		"double.nested": "⪢",
		"eq":            "≥",
		"eq.slant":      "⩾",
		"eq.lt":         "⋛",
		"eq.not":        "≱",
		"equiv":         "≧",
		"lt":            "≷",
		"lt.not":        "≹",
		"neq":           "⪈",
		"napprox":       "⪊",
		"nequiv":        "≩",
		"not":           "≯",
		"ntilde":        "⋧",
		"tilde":         "≳",
		"tilde.not":     "≵",
		"tri":           "⊳",
		"tri.eq":        "⊵",
		"tri.eq.not":    "⋭",
		"tri.not":       "⋫",
		"triple":        "⋙",
		"triple.nested": "⫸",
	})

	m.Symbols["lt"] = newSymbol("lt", map[string]string{
		"":              "<",
		"o":             "⧀",
		"dot":           "⋖",
		"approx":        "⪅",
		"arc":           "⪦",
		"arc.eq":        "⪨",
		"double":        "≪",
		"double.nested": "⪡",
		"eq":            "≤",
		"eq.slant":      "⩽",
		"eq.gt":         "⋚",
		"eq.not":        "≰",
		"equiv":         "≦",
		"gt":            "≶",
		"gt.not":        "≸",
		"neq":           "⪇",
		"napprox":       "⪉",
		"nequiv":        "≨",
		"not":           "≮",
		"ntilde":        "⋦",
		"tilde":         "≲",
		"tilde.not":     "≴",
		"tri":           "⊲",
		"tri.eq":        "⊴",
		"tri.eq.not":    "⋬",
		"tri.not":       "⋪",
		"triple":        "⋘",
		"triple.nested": "⫷",
	})

	m.Symbols["approx"] = newSymbol("approx", map[string]string{
		"":    "≈",
		"eq":  "≊",
		"not": "≉",
	})

	m.Symbols["prec"] = newSymbol("prec", map[string]string{
		"":             "≺",
		"approx":       "⪷",
		"curly.eq":     "≼",
		"curly.eq.not": "⋠",
		"double":       "⪻",
		"eq":           "⪯",
		"equiv":        "⪳",
		"napprox":      "⪹",
		"neq":          "⪱",
		"nequiv":       "⪵",
		"not":          "⊀",
		"ntilde":       "⋨",
		"tilde":        "≾",
	})

	m.Symbols["succ"] = newSymbol("succ", map[string]string{
		"":             "≻",
		"approx":       "⪸",
		"curly.eq":     "≽",
		"curly.eq.not": "⋡",
		"double":       "⪼",
		"eq":           "⪰",
		"equiv":        "⪴",
		"napprox":      "⪺",
		"neq":          "⪲",
		"nequiv":       "⪶",
		"not":          "⊁",
		"ntilde":       "⋩",
		"tilde":        "≿",
	})

	m.Symbols["equiv"] = newSymbol("equiv", map[string]string{
		"":    "≡",
		"not": "≢",
	})

	m.Symbols["smt"] = newSymbol("smt", map[string]string{
		"":   "⪪",
		"eq": "⪬",
	})

	m.Symbols["lat"] = newSymbol("lat", map[string]string{
		"":   "⪫",
		"eq": "⪭",
	})

	m.Symbols["prop"] = singleSymbol("prop", "∝")
	m.Symbols["original"] = singleSymbol("original", "⊶")
	m.Symbols["image"] = singleSymbol("image", "⊷")

	m.Symbols["asymp"] = newSymbol("asymp", map[string]string{
		"":    "≍",
		"not": "≭",
	})

	// Set theory
	m.Symbols["emptyset"] = newSymbol("emptyset", map[string]string{
		"":        "∅",
		"arrow.r": "⦳",
		"arrow.l": "⦴",
		"bar":     "⦱",
		"circle":  "⦲",
		"rev":     "⦰",
	})

	m.Symbols["nothing"] = newSymbol("nothing", map[string]string{
		"":        "∅",
		"arrow.r": "⦳",
		"arrow.l": "⦴",
		"bar":     "⦱",
		"circle":  "⦲",
		"rev":     "⦰",
	})

	m.Symbols["without"] = singleSymbol("without", "∖")
	m.Symbols["complement"] = singleSymbol("complement", "∁")

	m.Symbols["in"] = newSymbol("in", map[string]string{
		"":          "∈",
		"not":       "∉",
		"rev":       "∋",
		"rev.not":   "∌",
		"rev.small": "∍",
		"small":     "∊",
	})

	m.Symbols["subset"] = newSymbol("subset", map[string]string{
		"":           "⊂",
		"approx":     "⫉",
		"closed":     "⫏",
		"closed.eq":  "⫑",
		"dot":        "⪽",
		"double":     "⋐",
		"eq":         "⊆",
		"eq.dot":     "⫃",
		"eq.not":     "⊈",
		"eq.sq":      "⊑",
		"eq.sq.not":  "⋢",
		"equiv":      "⫅",
		"neq":        "⊊",
		"nequiv":     "⫋",
		"not":        "⊄",
		"plus":       "⪿",
		"sq":         "⊏",
		"sq.neq":     "⋤",
		"tilde":      "⫇",
		"times":      "⫁",
	})

	m.Symbols["supset"] = newSymbol("supset", map[string]string{
		"":           "⊃",
		"approx":     "⫊",
		"closed":     "⫐",
		"closed.eq":  "⫒",
		"dot":        "⪾",
		"double":     "⋑",
		"eq":         "⊇",
		"eq.dot":     "⫄",
		"eq.not":     "⊉",
		"eq.sq":      "⊒",
		"eq.sq.not":  "⋣",
		"equiv":      "⫆",
		"neq":        "⊋",
		"nequiv":     "⫌",
		"not":        "⊅",
		"plus":       "⫀",
		"sq":         "⊐",
		"sq.neq":     "⋥",
		"tilde":      "⫈",
		"times":      "⫂",
	})

	m.Symbols["union"] = newSymbol("union", map[string]string{
		"":          "∪",
		"serif":     "∪",
		"arrow":     "⊌",
		"big":       "⋃",
		"dot":       "⊍",
		"dot.big":   "⨃",
		"double":    "⋓",
		"minus":     "⩁",
		"or":        "⩅",
		"plus":      "⊎",
		"plus.big":  "⨄",
		"sq":        "⊔",
		"sq.serif":  "⊔",
		"sq.big":    "⨆",
		"sq.double": "⩏",
	})

	m.Symbols["inter"] = newSymbol("inter", map[string]string{
		"":          "∩",
		"serif":     "∩",
		"and":       "⩄",
		"big":       "⋂",
		"dot":       "⩀",
		"double":    "⋒",
		"sq":        "⊓",
		"sq.serif":  "⊓",
		"sq.big":    "⨅",
		"sq.double": "⩎",
	})

	// Calculus
	m.Symbols["infinity"] = newSymbol("infinity", map[string]string{
		"":           "∞",
		"bar":        "⧞",
		"incomplete": "⧜",
		"tie":        "⧝",
	})

	m.Symbols["oo"] = singleSymbol("oo", "∞")
	m.Symbols["partial"] = singleSymbol("partial", "∂")
	m.Symbols["gradient"] = singleSymbol("gradient", "∇")
	m.Symbols["nabla"] = singleSymbol("nabla", "∇")

	m.Symbols["sum"] = newSymbol("sum", map[string]string{
		"":         "∑",
		"integral": "⨋",
	})

	m.Symbols["product"] = newSymbol("product", map[string]string{
		"":   "∏",
		"co": "∐",
	})

	m.Symbols["integral"] = newSymbol("integral", map[string]string{
		"":           "∫",
		"arrow.hook": "⨗",
		"ccw":        "⨑",
		"cont":       "∮",
		"cont.ccw":   "∳",
		"cont.cw":    "∲",
		"cw":         "∱",
		"dash":       "⨍",
		"dash.double":"⨎",
		"double":     "∬",
		"quad":       "⨌",
		"inter":      "⨙",
		"slash":      "⨏",
		"square":     "⨖",
		"surf":       "∯",
		"times":      "⨘",
		"triple":     "∭",
		"union":      "⨚",
		"vol":        "∰",
	})

	m.Symbols["laplace"] = singleSymbol("laplace", "∆")

	// Logic
	m.Symbols["forall"] = singleSymbol("forall", "∀")

	m.Symbols["exists"] = newSymbol("exists", map[string]string{
		"":    "∃",
		"not": "∄",
	})

	m.Symbols["top"] = singleSymbol("top", "⊤")
	m.Symbols["bot"] = singleSymbol("bot", "⊥")
	m.Symbols["not"] = singleSymbol("not", "¬")

	m.Symbols["and"] = newSymbol("and", map[string]string{
		"":       "∧",
		"big":    "⋀",
		"curly":  "⋏",
		"dot":    "⟑",
		"double": "⩓",
	})

	m.Symbols["or"] = newSymbol("or", map[string]string{
		"":       "∨",
		"big":    "⋁",
		"curly":  "⋎",
		"dot":    "⟇",
		"double": "⩔",
	})

	m.Symbols["xor"] = newSymbol("xor", map[string]string{
		"":    "⊕",
		"big": "⨁",
	})

	m.Symbols["models"] = singleSymbol("models", "⊧")

	m.Symbols["forces"] = newSymbol("forces", map[string]string{
		"":    "⊩",
		"not": "⊮",
	})

	m.Symbols["therefore"] = singleSymbol("therefore", "∴")
	m.Symbols["because"] = singleSymbol("because", "∵")
	m.Symbols["qed"] = singleSymbol("qed", "∎")

	// Function and category theory
	m.Symbols["mapsto"] = newSymbol("mapsto", map[string]string{
		"":     "↦",
		"long": "⟼",
	})

	m.Symbols["compose"] = newSymbol("compose", map[string]string{
		"":  "∘",
		"o": "⊚",
	})

	m.Symbols["convolve"] = newSymbol("convolve", map[string]string{
		"":  "∗",
		"o": "⊛",
	})

	m.Symbols["multimap"] = newSymbol("multimap", map[string]string{
		"":       "⊸",
		"double": "⧟",
	})

	// Game theory
	m.Symbols["tiny"] = singleSymbol("tiny", "⧾")
	m.Symbols["miny"] = singleSymbol("miny", "⧿")

	// Number theory
	m.Symbols["divides"] = newSymbol("divides", map[string]string{
		"":        "∣",
		"not":     "∤",
		"not.rev": "⫮",
		"struck":  "⟊",
	})

	// Algebra
	m.Symbols["wreath"] = singleSymbol("wreath", "≀")

	// Geometry
	m.Symbols["angle"] = newSymbol("angle", map[string]string{
		"":             "∠",
		"acute":        "⦟",
		"arc":          "∡",
		"arc.rev":      "⦛",
		"azimuth":      "⍼",
		"obtuse":       "⦦",
		"rev":          "⦣",
		"right":        "∟",
		"right.rev":    "⯾",
		"right.arc":    "⊾",
		"right.dot":    "⦝",
		"right.square": "⦜",
		"s":            "⦞",
		"spatial":      "⟀",
		"spheric":      "∢",
		"spheric.rev":  "⦠",
		"spheric.t":    "⦡",
	})

	m.Symbols["angzarr"] = singleSymbol("angzarr", "⍼")

	m.Symbols["parallel"] = newSymbol("parallel", map[string]string{
		"":                "∥",
		"struck":          "⫲",
		"o":               "⦷",
		"eq":              "⋕",
		"equiv":           "⩨",
		"not":             "∦",
		"slanted.eq":      "⧣",
		"slanted.eq.tilde":"⧤",
		"slanted.equiv":   "⧥",
		"tilde":           "⫳",
	})

	m.Symbols["perp"] = newSymbol("perp", map[string]string{
		"":  "⟂",
		"o": "⦹",
	})

	// Astronomical
	m.Symbols["earth"] = newSymbol("earth", map[string]string{
		"":    "🜨",
		"alt": "♁",
	})

	m.Symbols["jupiter"] = singleSymbol("jupiter", "♃")
	m.Symbols["mars"] = singleSymbol("mars", "♂")
	m.Symbols["mercury"] = singleSymbol("mercury", "☿")

	m.Symbols["neptune"] = newSymbol("neptune", map[string]string{
		"":    "♆",
		"alt": "⯉",
	})

	m.Symbols["saturn"] = singleSymbol("saturn", "♄")
	m.Symbols["sun"] = singleSymbol("sun", "☉")

	m.Symbols["uranus"] = newSymbol("uranus", map[string]string{
		"":    "⛢",
		"alt": "♅",
	})

	m.Symbols["venus"] = singleSymbol("venus", "♀")

	// Miscellaneous Technical
	m.Symbols["diameter"] = singleSymbol("diameter", "⌀")

	m.Symbols["interleave"] = newSymbol("interleave", map[string]string{
		"":       "⫴",
		"big":    "⫼",
		"struck": "⫵",
	})

	m.Symbols["join"] = newSymbol("join", map[string]string{
		"":    "⨝",
		"r":   "⟖",
		"l":   "⟕",
		"l.r": "⟗",
	})

	m.Symbols["hourglass"] = newSymbol("hourglass", map[string]string{
		"stroked": "⧖",
		"filled":  "⧗",
	})

	m.Symbols["degree"] = singleSymbol("degree", "°")
	m.Symbols["smash"] = singleSymbol("smash", "⨳")

	m.Symbols["power"] = newSymbol("power", map[string]string{
		"standby": "⏻",
		"on":      "⏽",
		"off":     "⭘",
		"on.off":  "⏼",
		"sleep":   "⏾",
	})

	m.Symbols["smile"] = singleSymbol("smile", "⌣")
	m.Symbols["frown"] = singleSymbol("frown", "⌢")

	// Currency
	m.Symbols["afghani"] = singleSymbol("afghani", "؋")
	m.Symbols["baht"] = singleSymbol("baht", "฿")
	m.Symbols["bitcoin"] = singleSymbol("bitcoin", "₿")
	m.Symbols["cedi"] = singleSymbol("cedi", "₵")
	m.Symbols["cent"] = singleSymbol("cent", "¢")
	m.Symbols["currency"] = singleSymbol("currency", "¤")
	m.Symbols["dollar"] = singleSymbol("dollar", "$")
	m.Symbols["dong"] = singleSymbol("dong", "₫")
	m.Symbols["dorome"] = singleSymbol("dorome", "߾")
	m.Symbols["dram"] = singleSymbol("dram", "֏")
	m.Symbols["euro"] = singleSymbol("euro", "€")
	m.Symbols["guarani"] = singleSymbol("guarani", "₲")
	m.Symbols["hryvnia"] = singleSymbol("hryvnia", "₴")
	m.Symbols["kip"] = singleSymbol("kip", "₭")
	m.Symbols["lari"] = singleSymbol("lari", "₾")
	m.Symbols["lira"] = singleSymbol("lira", "₺")
	m.Symbols["manat"] = singleSymbol("manat", "₼")
	m.Symbols["naira"] = singleSymbol("naira", "₦")
	m.Symbols["pataca"] = singleSymbol("pataca", "$")
	m.Symbols["peso"] = newSymbol("peso", map[string]string{
		"":          "$",
		"philippine": "₱",
	})
	m.Symbols["pound"] = singleSymbol("pound", "£")
	m.Symbols["riel"] = singleSymbol("riel", "៛")
	m.Symbols["riyal"] = singleSymbol("riyal", "⃁")
	m.Symbols["ruble"] = singleSymbol("ruble", "₽")
	m.Symbols["rupee"] = newSymbol("rupee", map[string]string{
		"indian":  "₹",
		"generic": "₨",
		"tamil":   "௹",
		"wancho":  "𞋿",
	})
	m.Symbols["shekel"] = singleSymbol("shekel", "₪")
	m.Symbols["som"] = singleSymbol("som", "⃀")
	m.Symbols["taka"] = singleSymbol("taka", "৳")
	m.Symbols["taman"] = singleSymbol("taman", "߿")
	m.Symbols["tenge"] = singleSymbol("tenge", "₸")
	m.Symbols["togrog"] = singleSymbol("togrog", "₮")
	m.Symbols["won"] = singleSymbol("won", "₩")
	m.Symbols["yen"] = singleSymbol("yen", "¥")
	m.Symbols["yuan"] = singleSymbol("yuan", "¥")

	// Miscellaneous
	m.Symbols["ballot"] = newSymbol("ballot", map[string]string{
		"":           "☐",
		"cross":      "☒",
		"check":      "☑",
		"check.heavy":"🗹",
	})

	m.Symbols["checkmark"] = newSymbol("checkmark", map[string]string{
		"":      "✓",
		"light": "🗸",
		"heavy": "✔",
	})

	m.Symbols["crossmark"] = newSymbol("crossmark", map[string]string{
		"":      "✗",
		"heavy": "✘",
	})

	m.Symbols["floral"] = newSymbol("floral", map[string]string{
		"":  "❦",
		"l": "☙",
		"r": "❧",
	})

	m.Symbols["refmark"] = singleSymbol("refmark", "※")

	m.Symbols["cc"] = newSymbol("cc", map[string]string{
		"":       "🅭",
		"by":     "🅯",
		"nc":     "🄏",
		"nd":     "⊜",
		"public": "🅮",
		"sa":     "🄎",
		"zero":   "🄍",
	})

	m.Symbols["copyright"] = newSymbol("copyright", map[string]string{
		"":      "©",
		"sound": "℗",
	})

	m.Symbols["copyleft"] = singleSymbol("copyleft", "🄯")

	m.Symbols["trademark"] = newSymbol("trademark", map[string]string{
		"":           "™",
		"registered": "®",
		"service":    "℠",
	})

	m.Symbols["maltese"] = singleSymbol("maltese", "✠")

	m.Symbols["suit"] = newSymbol("suit", map[string]string{
		"club.filled":    "♣",
		"club.stroked":   "♧",
		"diamond.filled": "♦",
		"diamond.stroked":"♢",
		"heart.filled":   "♥",
		"heart.stroked":  "♡",
		"spade.filled":   "♠",
		"spade.stroked":  "♤",
	})

	// Music
	m.Symbols["note"] = newSymbol("note", map[string]string{
		"up":               "🎜",
		"down":             "🎝",
		"whole":            "𝅝",
		"half":             "𝅗𝅥",
		"quarter":          "𝅘𝅥",
		"quarter.alt":      "♩",
		"eighth":           "𝅘𝅥𝅮",
		"eighth.alt":       "♪",
		"eighth.beamed":    "♫",
		"sixteenth":        "𝅘𝅥𝅯",
		"sixteenth.beamed": "♬",
		"grace":            "𝆕",
		"grace.slash":      "𝆔",
	})

	m.Symbols["rest"] = newSymbol("rest", map[string]string{
		"whole":            "𝄻",
		"multiple":         "𝄺",
		"multiple.measure": "𝄩",
		"half":             "𝄼",
		"quarter":          "𝄽",
		"eighth":           "𝄾",
		"sixteenth":        "𝄿",
	})

	m.Symbols["natural"] = newSymbol("natural", map[string]string{
		"":  "♮",
		"t": "𝄮",
		"b": "𝄯",
	})

	m.Symbols["flat"] = newSymbol("flat", map[string]string{
		"":        "♭",
		"t":       "𝄬",
		"b":       "𝄭",
		"double":  "𝄫",
		"quarter": "𝄳",
	})

	m.Symbols["sharp"] = newSymbol("sharp", map[string]string{
		"":        "♯",
		"t":       "𝄰",
		"b":       "𝄱",
		"double":  "𝄪",
		"quarter": "𝄲",
	})

	// Shapes
	m.Symbols["bullet"] = newSymbol("bullet", map[string]string{
		"":           "•",
		"op":         "∙",
		"o":          "⦿",
		"stroked":    "◦",
		"stroked.o":  "⦾",
		"hole":       "◘",
		"hyph":       "⁃",
		"tri":        "‣",
		"l":          "⁌",
		"r":          "⁍",
	})

	m.Symbols["circle"] = newSymbol("circle", map[string]string{
		"stroked":       "○",
		"stroked.tiny":  "∘",
		"stroked.small": "⚬",
		"stroked.big":   "◯",
		"filled":        "●",
		"filled.tiny":   "⦁",
		"filled.small":  "∙",
		"filled.big":    "⬤",
		"dotted":        "◌",
	})

	m.Symbols["ellipse"] = newSymbol("ellipse", map[string]string{
		"stroked.h": "⬭",
		"stroked.v": "⬯",
		"filled.h":  "⬬",
		"filled.v":  "⬮",
	})

	m.Symbols["triangle"] = newSymbol("triangle", map[string]string{
		"stroked.t":         "△",
		"stroked.b":         "▽",
		"stroked.r":         "▷",
		"stroked.l":         "◁",
		"stroked.bl":        "◺",
		"stroked.br":        "◿",
		"stroked.tl":        "◸",
		"stroked.tr":        "◹",
		"stroked.small.t":   "▵",
		"stroked.small.b":   "▿",
		"stroked.small.r":   "▹",
		"stroked.small.l":   "◃",
		"stroked.rounded":   "🛆",
		"stroked.nested":    "⟁",
		"stroked.dot":       "◬",
		"filled.t":          "▲",
		"filled.b":          "▼",
		"filled.r":          "▶",
		"filled.l":          "◀",
		"filled.bl":         "◣",
		"filled.br":         "◢",
		"filled.tl":         "◤",
		"filled.tr":         "◥",
		"filled.small.t":    "▴",
		"filled.small.b":    "▾",
		"filled.small.r":    "▸",
		"filled.small.l":    "◂",
	})

	m.Symbols["square"] = newSymbol("square", map[string]string{
		"stroked":        "□",
		"stroked.tiny":   "▫",
		"stroked.small":  "◽",
		"stroked.medium": "◻",
		"stroked.big":    "⬜",
		"stroked.dotted": "⬚",
		"stroked.rounded":"▢",
		"filled":         "■",
		"filled.tiny":    "▪",
		"filled.small":   "◾",
		"filled.medium":  "◼",
		"filled.big":     "⬛",
	})

	m.Symbols["rect"] = newSymbol("rect", map[string]string{
		"stroked.h": "▭",
		"stroked.v": "▯",
		"filled.h":  "▬",
		"filled.v":  "▮",
	})

	m.Symbols["penta"] = newSymbol("penta", map[string]string{
		"stroked": "⬠",
		"filled":  "⬟",
	})

	m.Symbols["hexa"] = newSymbol("hexa", map[string]string{
		"stroked": "⬡",
		"filled":  "⬢",
	})

	m.Symbols["diamond"] = newSymbol("diamond", map[string]string{
		"stroked":        "◇",
		"stroked.small":  "⋄",
		"stroked.medium": "⬦",
		"stroked.dot":    "⟐",
		"filled":         "◆",
		"filled.medium":  "⬥",
		"filled.small":   "⬩",
	})

	m.Symbols["lozenge"] = newSymbol("lozenge", map[string]string{
		"stroked":        "◊",
		"stroked.small":  "⬫",
		"stroked.medium": "⬨",
		"filled":         "⧫",
		"filled.small":   "⬪",
		"filled.medium":  "⬧",
	})

	m.Symbols["parallelogram"] = newSymbol("parallelogram", map[string]string{
		"stroked": "▱",
		"filled":  "▰",
	})

	m.Symbols["star"] = newSymbol("star", map[string]string{
		"op":      "⋆",
		"stroked": "☆",
		"filled":  "★",
	})

	// Arrows - this is a large section
	m.Symbols["arrow"] = newSymbol("arrow", map[string]string{
		// Right arrows
		"r":                   "→",
		"r.long.bar":          "⟼",
		"r.bar":               "↦",
		"r.curve":             "⤷",
		"r.turn":              "⮎",
		"r.dashed":            "⇢",
		"r.dotted":            "⤑",
		"r.double":            "⇒",
		"r.double.bar":        "⤇",
		"r.double.long":       "⟹",
		"r.double.long.bar":   "⟾",
		"r.double.not":        "⇏",
		"r.double.struck":     "⤃",
		"r.filled":            "➡",
		"r.hook":              "↪",
		"r.long":              "⟶",
		"r.long.squiggly":     "⟿",
		"r.loop":              "↬",
		"r.not":               "↛",
		"r.quad":              "⭆",
		"r.squiggly":          "⇝",
		"r.stop":              "⇥",
		"r.stroked":           "⇨",
		"r.struck":            "⇸",
		"r.dstruck":           "⇻",
		"r.tail":              "↣",
		"r.tail.struck":       "⤔",
		"r.tail.dstruck":      "⤕",
		"r.tilde":             "⥲",
		"r.triple":            "⇛",
		"r.twohead":           "↠",
		"r.twohead.bar":       "⤅",
		"r.twohead.struck":    "⤀",
		"r.twohead.dstruck":   "⤁",
		"r.twohead.tail":      "⤖",
		"r.twohead.tail.struck":"⤗",
		"r.twohead.tail.dstruck":"⤘",
		"r.open":              "⇾",
		"r.wave":              "↝",
		// Left arrows
		"l":                   "←",
		"l.bar":               "↤",
		"l.curve":             "⤶",
		"l.turn":              "⮌",
		"l.dashed":            "⇠",
		"l.dotted":            "⬸",
		"l.double":            "⇐",
		"l.double.bar":        "⤆",
		"l.double.long":       "⟸",
		"l.double.long.bar":   "⟽",
		"l.double.not":        "⇍",
		"l.double.struck":     "⤂",
		"l.filled":            "⬅",
		"l.hook":              "↩",
		"l.long":              "⟵",
		"l.long.bar":          "⟻",
		"l.long.squiggly":     "⬳",
		"l.loop":              "↫",
		"l.not":               "↚",
		"l.quad":              "⭅",
		"l.squiggly":          "⇜",
		"l.stop":              "⇤",
		"l.stroked":           "⇦",
		"l.struck":            "⇷",
		"l.dstruck":           "⇺",
		"l.tail":              "↢",
		"l.tail.struck":       "⬹",
		"l.tail.dstruck":      "⬺",
		"l.tilde":             "⭉",
		"l.triple":            "⇚",
		"l.twohead":           "↞",
		"l.twohead.bar":       "⬶",
		"l.twohead.struck":    "⬴",
		"l.twohead.dstruck":   "⬵",
		"l.twohead.tail":      "⬻",
		"l.twohead.tail.struck":"⬼",
		"l.twohead.tail.dstruck":"⬽",
		"l.open":              "⇽",
		"l.wave":              "↜",
		// Up arrows
		"t":                   "↑",
		"t.bar":               "↥",
		"t.curve":             "⤴",
		"t.turn":              "⮍",
		"t.dashed":            "⇡",
		"t.double":            "⇑",
		"t.filled":            "⬆",
		"t.quad":              "⟰",
		"t.stop":              "⤒",
		"t.stroked":           "⇧",
		"t.struck":            "⤉",
		"t.dstruck":           "⇞",
		"t.triple":            "⤊",
		"t.twohead":           "↟",
		// Down arrows
		"b":                   "↓",
		"b.bar":               "↧",
		"b.curve":             "⤵",
		"b.turn":              "⮏",
		"b.dashed":            "⇣",
		"b.double":            "⇓",
		"b.filled":            "⬇",
		"b.quad":              "⟱",
		"b.stop":              "⤓",
		"b.stroked":           "⇩",
		"b.struck":            "⤈",
		"b.dstruck":           "⇟",
		"b.triple":            "⤋",
		"b.twohead":           "↡",
		// Bidirectional
		"l.r":                 "↔",
		"l.r.double":          "⇔",
		"l.r.double.long":     "⟺",
		"l.r.double.not":      "⇎",
		"l.r.double.struck":   "⤄",
		"l.r.filled":          "⬌",
		"l.r.long":            "⟷",
		"l.r.not":             "↮",
		"l.r.stroked":         "⬄",
		"l.r.struck":          "⇹",
		"l.r.dstruck":         "⇼",
		"l.r.open":            "⇿",
		"l.r.wave":            "↭",
		"t.b":                 "↕",
		"t.b.double":          "⇕",
		"t.b.filled":          "⬍",
		"t.b.stroked":         "⇳",
		// Diagonal
		"tr":                  "↗",
		"tr.double":           "⇗",
		"tr.filled":           "⬈",
		"tr.hook":             "⤤",
		"tr.stroked":          "⬀",
		"br":                  "↘",
		"br.double":           "⇘",
		"br.filled":           "⬊",
		"br.hook":             "⤥",
		"br.stroked":          "⬂",
		"tl":                  "↖",
		"tl.double":           "⇖",
		"tl.filled":           "⬉",
		"tl.hook":             "⤣",
		"tl.stroked":          "⬁",
		"bl":                  "↙",
		"bl.double":           "⇙",
		"bl.filled":           "⬋",
		"bl.hook":             "⤦",
		"bl.stroked":          "⬃",
		"tl.br":               "⤡",
		"tr.bl":               "⤢",
		// Circular
		"ccw":                 "↺",
		"ccw.half":            "↶",
		"cw":                  "↻",
		"cw.half":             "↷",
		"zigzag":              "↯",
	})

	m.Symbols["arrows"] = newSymbol("arrows", map[string]string{
		"rr":      "⇉",
		"ll":      "⇇",
		"tt":      "⇈",
		"bb":      "⇊",
		"lr":      "⇆",
		"lr.stop": "↹",
		"rl":      "⇄",
		"tb":      "⇅",
		"bt":      "⇵",
		"rrr":     "⇶",
		"lll":     "⬱",
	})

	m.Symbols["arrowhead"] = newSymbol("arrowhead", map[string]string{
		"t": "⌃",
		"b": "⌄",
	})

	m.Symbols["harpoon"] = newSymbol("harpoon", map[string]string{
		"rt":      "⇀",
		"rt.bar":  "⥛",
		"rt.stop": "⥓",
		"rb":      "⇁",
		"rb.bar":  "⥟",
		"rb.stop": "⥗",
		"lt":      "↼",
		"lt.bar":  "⥚",
		"lt.stop": "⥒",
		"lb":      "↽",
		"lb.bar":  "⥞",
		"lb.stop": "⥖",
		"tl":      "↿",
		"tl.bar":  "⥠",
		"tl.stop": "⥘",
		"tr":      "↾",
		"tr.bar":  "⥜",
		"tr.stop": "⥔",
		"bl":      "⇃",
		"bl.bar":  "⥡",
		"bl.stop": "⥙",
		"br":      "⇂",
		"br.bar":  "⥝",
		"br.stop": "⥕",
		"lt.rt":   "⥎",
		"lb.rb":   "⥐",
		"lb.rt":   "⥋",
		"lt.rb":   "⥊",
		"tl.bl":   "⥑",
		"tr.br":   "⥏",
		"tl.br":   "⥍",
		"tr.bl":   "⥌",
	})

	m.Symbols["harpoons"] = newSymbol("harpoons", map[string]string{
		"rtrb":  "⥤",
		"blbr":  "⥥",
		"bltr":  "⥯",
		"lbrb":  "⥧",
		"ltlb":  "⥢",
		"ltrb":  "⇋",
		"ltrt":  "⥦",
		"rblb":  "⥩",
		"rtlb":  "⇌",
		"rtlt":  "⥨",
		"tlbr":  "⥮",
		"tltr":  "⥣",
	})

	m.Symbols["tack"] = newSymbol("tack", map[string]string{
		"r":          "⊢",
		"r.not":      "⊬",
		"r.long":     "⟝",
		"r.short":    "⊦",
		"r.double":   "⊨",
		"r.double.not":"⊭",
		"l":          "⊣",
		"l.long":     "⟞",
		"l.short":    "⫞",
		"l.double":   "⫤",
		"t":          "⊥",
		"t.big":      "⟘",
		"t.double":   "⫫",
		"t.short":    "⫠",
		"b":          "⊤",
		"b.big":      "⟙",
		"b.double":   "⫪",
		"b.short":    "⫟",
		"l.r":        "⟛",
	})

	// Lowercase Greek
	m.Symbols["alpha"] = singleSymbol("alpha", "α")
	m.Symbols["beta"] = newSymbol("beta", map[string]string{
		"":    "β",
		"alt": "ϐ",
	})
	m.Symbols["chi"] = singleSymbol("chi", "χ")
	m.Symbols["delta"] = singleSymbol("delta", "δ")
	m.Symbols["digamma"] = singleSymbol("digamma", "ϝ")
	m.Symbols["epsilon"] = newSymbol("epsilon", map[string]string{
		"":        "ε",
		"alt":     "ϵ",
		"alt.rev": "϶",
	})
	m.Symbols["eta"] = singleSymbol("eta", "η")
	m.Symbols["gamma"] = singleSymbol("gamma", "γ")
	m.Symbols["iota"] = newSymbol("iota", map[string]string{
		"":    "ι",
		"inv": "℩",
	})
	m.Symbols["kappa"] = newSymbol("kappa", map[string]string{
		"":    "κ",
		"alt": "ϰ",
	})
	m.Symbols["lambda"] = singleSymbol("lambda", "λ")
	m.Symbols["mu"] = singleSymbol("mu", "μ")
	m.Symbols["nu"] = singleSymbol("nu", "ν")
	m.Symbols["omega"] = singleSymbol("omega", "ω")
	m.Symbols["omicron"] = singleSymbol("omicron", "ο")
	m.Symbols["phi"] = newSymbol("phi", map[string]string{
		"":    "φ",
		"alt": "ϕ",
	})
	m.Symbols["pi"] = newSymbol("pi", map[string]string{
		"":    "π",
		"alt": "ϖ",
	})
	m.Symbols["psi"] = singleSymbol("psi", "ψ")
	m.Symbols["rho"] = newSymbol("rho", map[string]string{
		"":    "ρ",
		"alt": "ϱ",
	})
	m.Symbols["sigma"] = newSymbol("sigma", map[string]string{
		"":    "σ",
		"alt": "ς",
	})
	m.Symbols["tau"] = singleSymbol("tau", "τ")
	m.Symbols["theta"] = newSymbol("theta", map[string]string{
		"":    "θ",
		"alt": "ϑ",
	})
	m.Symbols["upsilon"] = singleSymbol("upsilon", "υ")
	m.Symbols["xi"] = singleSymbol("xi", "ξ")
	m.Symbols["zeta"] = singleSymbol("zeta", "ζ")

	// Uppercase Greek
	m.Symbols["Alpha"] = singleSymbol("Alpha", "Α")
	m.Symbols["Beta"] = singleSymbol("Beta", "Β")
	m.Symbols["Chi"] = singleSymbol("Chi", "Χ")
	m.Symbols["Delta"] = singleSymbol("Delta", "Δ")
	m.Symbols["Digamma"] = singleSymbol("Digamma", "Ϝ")
	m.Symbols["Epsilon"] = singleSymbol("Epsilon", "Ε")
	m.Symbols["Eta"] = singleSymbol("Eta", "Η")
	m.Symbols["Gamma"] = singleSymbol("Gamma", "Γ")
	m.Symbols["Iota"] = singleSymbol("Iota", "Ι")
	m.Symbols["Kappa"] = singleSymbol("Kappa", "Κ")
	m.Symbols["Lambda"] = singleSymbol("Lambda", "Λ")
	m.Symbols["Mu"] = singleSymbol("Mu", "Μ")
	m.Symbols["Nu"] = singleSymbol("Nu", "Ν")
	m.Symbols["Omega"] = newSymbol("Omega", map[string]string{
		"":    "Ω",
		"inv": "℧",
	})
	m.Symbols["Omicron"] = singleSymbol("Omicron", "Ο")
	m.Symbols["Phi"] = singleSymbol("Phi", "Φ")
	m.Symbols["Pi"] = singleSymbol("Pi", "Π")
	m.Symbols["Psi"] = singleSymbol("Psi", "Ψ")
	m.Symbols["Rho"] = singleSymbol("Rho", "Ρ")
	m.Symbols["Sigma"] = singleSymbol("Sigma", "Σ")
	m.Symbols["Tau"] = singleSymbol("Tau", "Τ")
	m.Symbols["Theta"] = newSymbol("Theta", map[string]string{
		"":    "Θ",
		"alt": "ϴ",
	})
	m.Symbols["Upsilon"] = singleSymbol("Upsilon", "Υ")
	m.Symbols["Xi"] = singleSymbol("Xi", "Ξ")
	m.Symbols["Zeta"] = singleSymbol("Zeta", "Ζ")

	// Cyrillic
	m.Symbols["sha"] = singleSymbol("sha", "ш")
	m.Symbols["Sha"] = singleSymbol("Sha", "Ш")

	// Hebrew
	m.Symbols["aleph"] = singleSymbol("aleph", "א")
	m.Symbols["beth"] = singleSymbol("beth", "ב")
	m.Symbols["gimel"] = singleSymbol("gimel", "ג")
	m.Symbols["daleth"] = singleSymbol("daleth", "ד")

	// Double-struck letters
	m.Symbols["AA"] = singleSymbol("AA", "𝔸")
	m.Symbols["BB"] = singleSymbol("BB", "𝔹")
	m.Symbols["CC"] = singleSymbol("CC", "ℂ")
	m.Symbols["DD"] = singleSymbol("DD", "𝔻")
	m.Symbols["EE"] = singleSymbol("EE", "𝔼")
	m.Symbols["FF"] = singleSymbol("FF", "𝔽")
	m.Symbols["GG"] = singleSymbol("GG", "𝔾")
	m.Symbols["HH"] = singleSymbol("HH", "ℍ")
	m.Symbols["II"] = singleSymbol("II", "𝕀")
	m.Symbols["JJ"] = singleSymbol("JJ", "𝕁")
	m.Symbols["KK"] = singleSymbol("KK", "𝕂")
	m.Symbols["LL"] = singleSymbol("LL", "𝕃")
	m.Symbols["MM"] = singleSymbol("MM", "𝕄")
	m.Symbols["NN"] = singleSymbol("NN", "ℕ")
	m.Symbols["OO"] = singleSymbol("OO", "𝕆")
	m.Symbols["PP"] = singleSymbol("PP", "ℙ")
	m.Symbols["QQ"] = singleSymbol("QQ", "ℚ")
	m.Symbols["RR"] = singleSymbol("RR", "ℝ")
	m.Symbols["SS"] = singleSymbol("SS", "𝕊")
	m.Symbols["TT"] = singleSymbol("TT", "𝕋")
	m.Symbols["UU"] = singleSymbol("UU", "𝕌")
	m.Symbols["VV"] = singleSymbol("VV", "𝕍")
	m.Symbols["WW"] = singleSymbol("WW", "𝕎")
	m.Symbols["XX"] = singleSymbol("XX", "𝕏")
	m.Symbols["YY"] = singleSymbol("YY", "𝕐")
	m.Symbols["ZZ"] = singleSymbol("ZZ", "ℤ")

	// Miscellaneous letter-likes
	m.Symbols["angstrom"] = singleSymbol("angstrom", "Å")
	m.Symbols["ell"] = singleSymbol("ell", "ℓ")
	m.Symbols["pee"] = singleSymbol("pee", "℘")
	m.Symbols["planck"] = singleSymbol("planck", "ħ")
	m.Symbols["Re"] = singleSymbol("Re", "ℜ")
	m.Symbols["Im"] = singleSymbol("Im", "ℑ")

	m.Symbols["dotless"] = newSymbol("dotless", map[string]string{
		"i": "ı",
		"j": "ȷ",
	})

	// Miscellany
	m.Symbols["die"] = newSymbol("die", map[string]string{
		"six":   "⚅",
		"five":  "⚄",
		"four":  "⚃",
		"three": "⚂",
		"two":   "⚁",
		"one":   "⚀",
	})

	m.Symbols["errorbar"] = newSymbol("errorbar", map[string]string{
		"square.stroked":  "⧮",
		"square.filled":   "⧯",
		"diamond.stroked": "⧰",
		"diamond.filled":  "⧱",
		"circle.stroked":  "⧲",
		"circle.filled":   "⧳",
	})

	// Gender module
	genderMod := &Module{
		Name:    "gender",
		Symbols: make(map[string]*Symbol),
	}
	genderMod.Symbols["female"] = newSymbol("female", map[string]string{
		"":       "♀",
		"double": "⚢",
		"male":   "⚤",
	})
	genderMod.Symbols["intersex"] = singleSymbol("intersex", "⚥")
	genderMod.Symbols["male"] = newSymbol("male", map[string]string{
		"":         "♂",
		"double":   "⚣",
		"female":   "⚤",
		"stroke":   "⚦",
		"stroke.t": "⚨",
		"stroke.r": "⚩",
	})
	genderMod.Symbols["neuter"] = singleSymbol("neuter", "⚲")
	genderMod.Symbols["trans"] = singleSymbol("trans", "⚧")
	m.Submodules["gender"] = genderMod

	return m
}
