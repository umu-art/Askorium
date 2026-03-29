package quality

// Single-word noise tokens (lowercased).
var noiseTokens = map[string]struct{}{
	// EN
	"home": {}, "login": {}, "logout": {}, "signup": {}, "subscribe": {},
	"newsletter": {}, "cookie": {}, "cookies": {}, "copyright": {},
	"disclaimer": {}, "advertisement": {}, "share": {}, "tweet": {},
	"facebook": {}, "twitter": {}, "instagram": {}, "linkedin": {},
	"menu": {}, "navigation": {}, "sidebar": {}, "footer": {}, "header": {},
	"search": {}, "sitemap": {}, "faq": {}, "terms": {}, "conditions": {},
	"pinterest": {}, "youtube": {}, "tiktok": {}, "whatsapp": {},
	// RU
	"главная": {}, "войти": {}, "выйти": {}, "регистрация": {},
	"подписка": {}, "подписаться": {}, "реклама": {}, "навигация": {},
	"меню": {}, "поиск": {}, "карта": {}, "авторизация": {},
}

// Multi-word noise phrases (lowercased).
var noisePhrases = map[string]struct{}{
	// EN
	"sign up": {}, "sign in": {}, "log in": {}, "log out": {},
	"read more": {}, "click here": {}, "learn more": {},
	"privacy policy": {}, "terms of service": {}, "terms of use": {},
	"all rights reserved": {}, "powered by": {}, "back to top": {},
	"cookie policy": {}, "accept cookies": {},
	// RU
	"все права защищены": {}, "политика конфиденциальности": {},
	"обратная связь": {}, "пользовательское соглашение": {},
	"условия использования": {},
}
