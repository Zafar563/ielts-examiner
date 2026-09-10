# IELTS Writing Examiner Telegram Bot 🤖📝

IELTS Writing insholarini (Task 1 / Task 2) qabul qilib, ularni qat'iy va xolis xalqaro mezonlar asosida 75 ballik tizimda baholovchi hamda xatolarini batafsil tahlil qiluvchi Telegram bot.

---

## 🎯 Baholash Mezonlari (0–75 Ballik Tizim)

Bot inshoni quyidagi 4 ta asosiy mezon bo'yicha baholaydi:
1. **T/R** — Task Response (Vazifaga to'liq javob berilganligi, g'oyalar rivoji)
2. **C/C** — Coherence & Cohesion (Mantiqiy izchillik, paragraflar tuzilishi, bog'lovchilar)
3. **G/A** — Grammar & Accuracy (Grammatik tuzilmalar xilma-xilligi va aniqligi)
4. **L/R** — Lexical Resource (Akademik so'z boyligi, kollokatsiyalar, aniqlik)

### 📊 Darajalar shkalasi:
- **65–75 = C1** (Kuchli nazorat, keng qamrovli lug'at, murakkab grammatika, tabiiy bog'lanish)
- **51–64 = B2** (Samarali muloqot, ammo rivojlanish yoki lug'atda ko'zga tashlanadigan chegaralar)
- **41–50 = B1** (Tushunarli, lekin sodda va cheklangan grammatika va leksika)

### 📋 Har bir tahlil natijasida:
- Mezonlar bo'yicha alohida ballar va umumiy o'rtacha ball (X/75)
- Har bir mezon bo'yicha nima sababdan ushbu ball qo'yilganligi
- **Xatolar jadvali**: Asl jumla | To'g'rilangan variant | Xato turi (Grammar, Vocabulary, Collocation, Style...) | Izoh
- **Why it is NOT a higher score**: Nima sababdan yuqoriroq ball berilmaganligi (2–4 sabab)
- **How to reach 70+**: 70+ ballga chiqish uchun 3–5 ta aniq amaliy o'zgartirishlar
- **Rewrite**: Inshongizning asl g'oyasini saqlagan holda 70–75 ballik akademik variant

---

## 🛠 Texnologiyalar

- **Dasturlash tili:** Go (Golang 1.23)
- **Bot Framework:** `github.com/go-telegram-bot-api/telegram-bot-api/v5`
- **Ma'lumotlar bazasi:** PostgreSQL 16
- **Migratsiya:** Avtomatik SQL migratsiyalar (`migrations/`)
- **AI Integratsiyasi:** Google Gemini API (yoki OpenAI GPT-4o)
- **Konteynerlash:** Docker & Docker Compose

---

## 🚀 O'rnatish va Ishga Tushirish

### 1. Repozitoriyni sozlash va .env faylini yaratish
`.env.example` nusxasini olib `.env` faylini yarating:
```bash
cp .env.example .env
```

`.env` faylini ochib o'z sozlamalaringizni kiriting:
```env
TELEGRAM_BOT_TOKEN=123456789:ABCdefGhIJKlmNoPQRstuVWXyz
LLM_PROVIDER=gemini
LLM_API_KEY=AIzaSy...Sizning_Gemini_API_Kalingiz
LLM_MODEL=gemini-1.5-flash
```
*(Eslatma: OpenAI ishlatmoqchi bo'lsangiz `LLM_PROVIDER=openai`, `LLM_MODEL=gpt-4o-mini` va OpenAI kalitingizni kiriting)*.

---

### 2. Docker orqali ishga tushirish (Tavsiya etiladi) 🐳

Faqat bitta buyruq bilan PostgreSQL va Botni konteynerda ishga tushiring:
```bash
docker compose up -d --build
```

Loglarni ko'rish:
```bash
docker compose logs -f bot
```

To'xtatish:
```bash
docker compose down
```

---

### 3. Mahalliy (Local) ishga tushirish (Docker-siz)

PostgreSQL bazasi kompyuteringizda ishlab turgan bo'lsa:
```bash
# Bog'liqliklarni yuklash
go mod download

# Dasturni ishga tushirish
go run cmd/bot/main.go
```

---

## 📁 Loyiha Strukturasi

```
ielts_bot/
├── cmd/
│   └── bot/
│       └── main.go               # Ilova boshlang'ich nuqtasi (Entrypoint)
├── internal/
│   ├── config/                   # Konfiguratsiyalar va .env
│   ├── database/                 # Postgres ulanish, migratsiya va repository
│   ├── bot/                      # Telegram bot menyulari, holatlar (FSM) va xabarlar
│   ├── llm/                      # Gemini / OpenAI mijozi va IELTS Examiner prompti
│   └── models/                   # Ma'lumotlar modellari
├── migrations/                   # SQL migratsiya fayllari (.up.sql / .down.sql)
├── Dockerfile                    # Multi-stage Docker qurilishi
├── docker-compose.yml            # Docker orkestratsiyasi
├── .env.example                  # Namuna konfiguratsiya fayli
└── .gitignore                    # Git istisnolari
```
