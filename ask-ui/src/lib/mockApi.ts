import type { Message, Source } from '@/types'
import { generateId } from '@/lib/utils'

interface MockResponse {
  content: string
  sources: Source[]
}

const MOCK_RESPONSES: MockResponse[] = [
  {
    content: `## Дедлайны сдачи курсовых работ

Согласно регламенту НИУ ВШЭ, сроки сдачи курсовых работ зависят от факультета и образовательной программы:

- **Весенний семестр** — как правило, **конец апреля — начало мая**
- **Осенний семестр** — как правило, **декабрь**, до начала сессии

Конкретные даты устанавливаются кафедрой и научным руководителем. Актуальные дедлайны публикуются:
1. В личном кабинете студента на [my.hse.ru](https://my.hse.ru)
2. В Telegram-канале вашей образовательной программы
3. На странице кафедры на сайте ВШЭ

> **Важно:** уточняйте сроки у научного руководителя — они могут отличаться от общих дат.`,
    sources: [
      { id: 1, title: 'Регламент выполнения курсовых работ — НИУ ВШЭ', url: 'https://www.hse.ru/studyspravka/kurs/', domain: 'hse.ru' },
      { id: 2, title: 'Личный кабинет студента HSE', url: 'https://my.hse.ru', domain: 'my.hse.ru' },
      { id: 3, title: 'Учебный офис НИУ ВШЭ', url: 'https://www.hse.ru/org/hse/uo/', domain: 'hse.ru' },
    ],
  },
  {
    content: `## Порядок заселения в общежитие НИУ ВШЭ

Для получения места в общежитии необходимо:

1. **Подать заявку** через личный кабинет на [my.hse.ru](https://my.hse.ru) в период приёма заявлений
2. **Дождаться решения** — места распределяются по балльной системе (иногородние, льготные категории, академические успехи)
3. **Заключить договор** с дирекцией общежития после получения положительного решения
4. **Заселиться** в установленные сроки, предъявив документы

### Требуемые документы
- Паспорт
- Студенческий билет
- Медицинская книжка (или справка о прохождении флюорографии)
- Фотографии (обычно 2 шт., формат 3×4)

Подробная информация о стоимости и условиях — на официальной странице студгородка.`,
    sources: [
      { id: 1, title: 'Студенческий городок НИУ ВШЭ — официальная страница', url: 'https://www.hse.ru/studcity/', domain: 'hse.ru' },
      { id: 2, title: 'Подача заявки на общежитие — my.hse.ru', url: 'https://my.hse.ru/dormitory', domain: 'my.hse.ru' },
      { id: 3, title: 'Правила проживания в студгородке', url: 'https://www.hse.ru/studcity/rules/', domain: 'hse.ru' },
    ],
  },
  {
    content: `## Академический отпуск в НИУ ВШЭ

Академический отпуск предоставляется по следующим основаниям:

- **По состоянию здоровья** — при наличии медицинских показаний
- **По семейным обстоятельствам** — уход за ребёнком до 3 лет, уход за больным родственником
- **По призыву на военную службу**
- **По иным обстоятельствам** — решается индивидуально

### Процедура оформления

1. Написать **заявление** на имя директора образовательной программы
2. Приложить **подтверждающие документы** (справки, свидетельства и т.д.)
3. Передать документы в **учебный офис**
4. Получить приказ об академическом отпуске

Срок рассмотрения заявления — **до 10 рабочих дней**.

> Во время академического отпуска студент сохраняет статус и место в общежитии (в случае медицинского отпуска).`,
    sources: [
      { id: 1, title: 'Академический отпуск — правила НИУ ВШЭ', url: 'https://www.hse.ru/studyspravka/academ/', domain: 'hse.ru' },
      { id: 2, title: 'Учебный регламент НИУ ВШЭ', url: 'https://www.hse.ru/docs/regulations/', domain: 'hse.ru' },
      { id: 3, title: 'Контакты учебного офиса', url: 'https://www.hse.ru/org/hse/uo/contacts/', domain: 'hse.ru' },
    ],
  },
  {
    content: `## Стипендии в НИУ ВШЭ

НИУ ВШЭ предлагает несколько видов стипендий:

### Академическая стипендия
- Назначается студентам, сдавшим сессию **без троек**
- Размер зависит от курса и программы
- Выплачивается **ежемесячно**

### Повышенная академическая стипендия
- Для студентов с **выдающимися успехами** (топ по GPA + активность)
- Назначается по конкурсу раз в семестр

### Социальная стипендия
- Для студентов из льготных категорий (сироты, инвалиды, малообеспеченные)
- Назначается независимо от успеваемости

### Именные стипендии
- Стипендия Правительства РФ
- Стипендия Президента РФ
- Корпоративные стипендии партнёров ВШЭ

Все заявки подаются через [my.hse.ru](https://my.hse.ru) в период стипендиальной кампании.`,
    sources: [
      { id: 1, title: 'Стипендии и льготы — НИУ ВШЭ', url: 'https://www.hse.ru/studyspravka/stipend/', domain: 'hse.ru' },
      { id: 2, title: 'Стипендиальная комиссия — my.hse.ru', url: 'https://my.hse.ru/scholarships', domain: 'my.hse.ru' },
      { id: 3, title: 'Социальная поддержка студентов', url: 'https://www.hse.ru/studyspravka/social/', domain: 'hse.ru' },
    ],
  },
  {
    content: `## Ответ на ваш вопрос

По данному запросу на сайте НИУ ВШЭ найдена следующая информация:

На официальном портале ВШЭ размещены подробные материалы по данной теме. Рекомендуем обратиться к следующим разделам:

- **Официальный сайт**: [hse.ru](https://www.hse.ru) — главный источник актуальной информации
- **Личный кабинет**: [my.hse.ru](https://my.hse.ru) — персонализированные данные и сервисы
- **Учебный офис**: прямое обращение для уточнения деталей

Если ответ недостаточно полный, попробуйте переформулировать вопрос или обратитесь в учебный офис вашего факультета.`,
    sources: [
      { id: 1, title: 'НИУ ВШЭ — Официальный сайт', url: 'https://www.hse.ru', domain: 'hse.ru' },
      { id: 2, title: 'Личный кабинет студента', url: 'https://my.hse.ru', domain: 'my.hse.ru' },
      { id: 3, title: 'Справочник студента НИУ ВШЭ', url: 'https://www.hse.ru/studyspravka/', domain: 'hse.ru' },
    ],
  },
]

/**
 * Simulates a backend API call with realistic delay.
 * This layer will be replaced with actual HTTP calls once backend contracts are defined.
 * The interface is designed to mirror the expected backend response shape.
 */
export async function sendMessage(query: string): Promise<Message> {
  // Simulate network latency (1.5–3s to feel realistic)
  const delay = 1500 + Math.random() * 1500
  await new Promise(resolve => setTimeout(resolve, delay))

  // Simple keyword routing to pick a relevant mock response
  const lowerQuery = query.toLowerCase()
  let response: MockResponse

  if (lowerQuery.includes('курсов') || lowerQuery.includes('дедлайн') || lowerQuery.includes('сдач')) {
    response = MOCK_RESPONSES[0]
  } else if (lowerQuery.includes('общежит') || lowerQuery.includes('заселен') || lowerQuery.includes('студгородок')) {
    response = MOCK_RESPONSES[1]
  } else if (lowerQuery.includes('академич') || lowerQuery.includes('отпуск')) {
    response = MOCK_RESPONSES[2]
  } else if (lowerQuery.includes('стипенд')) {
    response = MOCK_RESPONSES[3]
  } else {
    response = MOCK_RESPONSES[4]
  }

  return {
    id: generateId(),
    role: 'assistant',
    content: response.content,
    sources: response.sources,
    timestamp: new Date(),
  }
}
