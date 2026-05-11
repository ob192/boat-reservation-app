export const MESSAGES = {
  // Navigation
  nav: {
    date: 'Дата',
    time: 'Час',
    boats: 'Човни',
    details: 'Деталі',
    done: 'Готово',
  },

  // Calendar
  calendar: {
    title: 'Виберіть дату',
    subtitle: 'Оберіть бажаний день на воді',
    prevMonth: 'Попередній місяць',
    nextMonth: 'Наступний місяць',
    available: 'Доступно',
    limited: 'Обмежено',
    booked: 'Заброньовано',
    days: ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Нд'],
    months: [
      'Січень', 'Лютий', 'Березень', 'Квітень', 'Травень', 'Червень',
      'Липень', 'Серпень', 'Вересень', 'Жовтень', 'Листопад', 'Грудень',
    ],
  },

  // Time slots
  time: {
    title: 'Виберіть час',
    subtitle: 'Оберіть час відправлення',
    fullTag: 'Зайнято',
    noSlots: 'Місць немає',
    slotsAvailable: (n: number) =>
      `${n} з 15 вільно`,
    periods: {
      '08:00': 'Ранок',
      '11:00': 'Пізній ранок',
      '15:00': 'День',
      '19:00': 'Вечір',
    } as Record<string, string>,
  },

  // Boats
  boats: {
    title: 'Виберіть човни',
    bigSection: 'Великі човни',
    compactSection: 'Компактні човни',
    bigName: 'Великий човен',
    mediumName: 'Середній човен',
    perBoat: 'за човен',
    children: 'Додати дітей',
    childrenBadge: '−50%',
    childrenDesc: 'Лише на великих човнах · до 40 кг · половина ціни',
    capacityWarn: '⚠️ Недостатньо місць у цьому часовому слоті. Будь ласка, зменшіть кількість.',
    decrease: 'Зменшити',
    increase: 'Збільшити',
    slotsAvailable: (n: number) =>
      `${n} ${n === 1 ? 'місце доступне' : n < 5 ? 'місця доступні' : 'місць доступно'}`,
  },

  // Summary
  summary: {
    selected: 'Обрано',
    bigLabel: 'Великий',
    mediumLabel: 'Середній',
    childLabel: 'Дитина',
    none: '—',
  },

  // Details form
  details: {
    title: 'Ваші дані',
    subtitle: 'Інформація для підтвердження',
    firstName: "Ім'я",
    lastName: 'Прізвище',
    email: 'Електронна пошта',
    emailReadonly: 'Пов\'язано з вашим обліковим записом',
    phone: 'Телефон (необов\'язково)',
    phonePlaceholder: '+380...',
    proceedToPayment: 'Перейти до оплати',
  },

  // Buttons
  buttons: {
    continue: 'Продовжити',
    back: 'Назад',
    book: 'Забронювати',
    newBooking: 'Зробити ще одне бронювання',
  },

  // Processing
  processing: {
    title: 'Обробка оплати...',
    subtitle: 'Будь ласка, зачекайте. Не закривайте цю сторінку.',
  },

  // Success
  success: {
    title: 'Готово!',
    message: 'Ваше бронювання підтверджено. До зустрічі на воді — деталі надішлемо на вашу електронну пошту.',
    dateLabel: 'Дата',
    departureLabel: 'Відправлення',
    bigBoatsLabel: 'Великі човни',
    mediumBoatsLabel: 'Середні човни',
    childrenLabel: 'Діти',
    totalLabel: 'Усього',
  },

  // Errors
  errors: {
    slotTaken: 'Цей часовий слот щойно зайняли. Будь ласка, оберіть інший.',
    bookingFailed: 'Не вдалося створити бронювання. Спробуйте ще раз.',
    backendUnavailable: 'Сервіс тимчасово недоступний. Спробуйте пізніше.',
    paymentFailed: 'Оплата не пройшла.',
    bookingExpired: 'Час бронювання вийшов.',
  },

  // Header
  header: {
    title: 'Harbour & Wave',
    subtitle: 'Бронювання човнів',
    season: 'Сезон 2025',
  },
};

export const PRICES = { big: 35, medium: 20, child: 17.5 };
export const MAX_SLOTS = 15;
export const TIME_OPTIONS = ['08:00', '11:00', '15:00', '19:00'] as const;
