package scoring

// English function words for stopword density gating.
// Prepositions, conjunctions, pronouns, articles, auxiliary verbs.
var enStopwords = []string{
	"a", "about", "above", "after", "again", "against", "all", "am", "an",
	"and", "any", "are", "as", "at",
	"be", "because", "been", "before", "being", "below", "between", "both",
	"but", "by",
	"can", "could",
	"did", "do", "does", "doing", "down", "during",
	"each", "even",
	"few", "for", "from",
	"had", "has", "have", "having", "he", "her", "here", "hers", "herself",
	"him", "himself", "his", "how",
	"i", "if", "in", "into", "is", "it", "its", "itself",
	"just",
	"me", "might", "more", "most", "must", "my", "myself",
	"no", "nor", "not", "now",
	"of", "off", "on", "once", "only", "or", "other", "our", "ours",
	"ourselves", "out", "over", "own",
	"same", "she", "should", "so", "some", "such",
	"than", "that", "the", "their", "theirs", "them", "themselves", "then",
	"there", "these", "they", "this", "those", "through", "to", "too",
	"under", "until", "up", "us",
	"very",
	"was", "we", "were", "what", "when", "where", "which", "while", "who",
	"whom", "why", "will", "with", "would",
	"you", "your", "yours", "yourself", "yourselves",
}

// Russian function words for stopword density gating.
// Source: filtered from the project's Russian stopword list, keeping only
// prepositions, conjunctions, pronouns, particles, and auxiliary verb forms.
// Content words (nouns, main verbs, adjectives) are excluded — they appear
// in both body text and noise (addresses, nav labels), destroying separation.
//
// 240 words.
var ruStopwords = []string{
	"а", "без", "более", "больше", "будем", "будет", "будете", "будешь",
	"будто", "буду", "будут", "будь", "бы", "был", "была", "были", "было",
	"быть",
	"в", "вам", "вами", "вас", "ваш", "ваша", "ваше", "ваши", "ведь", "во",
	"вот", "все", "всего", "всем", "всеми", "всему", "всех", "всею", "всю",
	"вся", "всё", "вы",
	"где",
	"да", "даже", "для", "до",
	"его", "ее", "ей", "ему", "если", "есть", "еще", "ещё", "ею", "её",
	"ж", "же",
	"за", "затем", "зато", "зачем",
	"и", "из", "или", "им", "именно", "ими", "иногда", "их",
	"к", "каждая", "каждое", "каждые", "каждый", "кажется", "как", "какая",
	"какой", "кем", "когда", "кого", "ком", "кому", "которая", "которого",
	"которой", "которые", "который", "которых", "кроме", "кто", "куда",
	"ли", "лишь",
	"между", "менее", "меньше", "меня", "мне", "мной", "мною", "мог", "могу",
	"могут", "может", "можно", "мои", "мой", "моя", "моё", "мы",
	"на", "над", "надо", "наиболее", "нам", "нами", "нас", "наш", "наша",
	"наше", "наши", "не", "него", "нее", "ней", "нельзя", "нем", "нему",
	"нет", "нею", "неё", "ни", "ним", "ними", "них", "но", "ну",
	"о", "об", "оба", "однако", "он", "она", "они", "оно", "от",
	"перед", "по", "под", "пока", "после", "потом", "потому", "почему",
	"почти", "при", "про", "просто", "против",
	"с", "сам", "сама", "сами", "самим", "самими", "самих", "само", "самого",
	"самой", "самом", "самому", "саму", "самый", "свое", "своего", "своей",
	"свои", "своих", "свой", "свою", "себе", "себя", "сейчас", "сих", "со",
	"собой", "собою", "суть",
	"та", "так", "такая", "также", "таки", "такие", "такое", "такой", "те",
	"тебе", "тебя", "тем", "теми", "теперь", "тех", "то", "тобой", "тобою",
	"тогда", "того", "тоже", "только", "том", "тому", "тот", "тою", "ту", "ты",
	"у", "уж", "уже",
	"чего", "чем", "чему", "через", "что", "чтоб", "чтобы",
	"эта", "эти", "этим", "этими", "этих", "это", "этого", "этой", "этом",
	"этому", "этот", "эту",
	"я",
}

func getStopwords() map[string]struct{} {
	set := make(map[string]struct{}, len(ruStopwords)+len(enStopwords))
	for _, w := range ruStopwords {
		set[w] = struct{}{}
	}
	for _, w := range enStopwords {
		set[w] = struct{}{}
	}
	return set
}
