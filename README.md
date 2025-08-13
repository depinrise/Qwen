# Qwen - Natural Conversational Telegram AI Bot

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blueviolet.svg)](https://www.gnu.org/licenses/agpl-3.0)

Bot Telegram sederhana yang menggunakan model AI Qwen dari Alibaba Cloud untuk percakapan interaktif.

## Fitur

- 🤖 Respons AI menggunakan model Qwen MT-Turbo
- 💬 Percakapan natural dengan pengguna
- ⚡ **Streaming Response** - Respons real-time dengan incremental streaming dari API
- 🔄 **WebSocket Support** - Interface web untuk testing streaming  
- 💬 **Real-time Typing Effect** - Melihat respons AI muncul word-by-word seperti typing
- 🗄️ **PolarDB MySQL Integration** - Riwayat percakapan tersimpan untuk konteks yang lebih baik
- 🧠 **Conversation Context** - AI mengingat percakapan sebelumnya untuk respons yang lebih relevan  
- 🧭 **Dynamic Memory System** - Bot mengingat informasi personal user secara permanen dengan LLM-based management
- 🎯 **Smart Information Extraction** - Ekstraksi otomatis informasi personal tanpa regex, menggunakan AI contextual analysis
- 🔄 **Memory-Enhanced Responses** - Personalisasi respons berdasarkan memory yang tersimpan dan dikelola secara dinamis
- 💬 **Natural Conversational AI** - Personality yang warm, adaptif, dan genuinely helpful
- 🎯 **Optimized Parameters** - Temperature 0.8 & Top-P 0.95 untuk respons yang natural
- 🚀 **High Performance** - Optimized streaming tanpa complex parsing
- 🔧 Command `/start`, `/help`, dan `/resetmemory`
- 🐳 Containerized dengan Docker
- 📦 Struktur kode modular

## Prasyarat

- API Key dari [Alibaba Cloud Model Studio](https://www.alibabacloud.com/help/en/model-studio/get-api-key)
- Bot Token dari [@BotFather](https://t.me/BotFather) di Telegram
- **PolarDB MySQL 8.0** dari Alibaba Cloud (opsional untuk conversation history)
- Docker dan Docker Compose (untuk deployment)
- Go 1.21+ (untuk development)

## Instalasi dan Setup

### 1. Clone Repository

```bash
git clone <repository-url>
cd Qwen
```

### 2. Setup Environment Variables

Salin file konfigurasi dan isi dengan kredensial Anda:

```bash
cp env.example .env
```

Edit file `.env`:

```env
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
DASHSCOPE_API_KEY=your_dashscope_api_key_here
DASHSCOPE_BASE_URL=https://dashscope-intl.aliyuncs.com/compatible-mode/v1
AI_MODEL=qwen-mt-turbo
HTTP_PORT=8080
DATABASE_DSN=user:password@tcp(your-polardb-host:3306)/telegram_bot?charset=utf8mb4&parseTime=True&loc=Local
```

### 3. Setup Database (Opsional)

Jika menggunakan PolarDB MySQL:

```bash
# 1. Buat database dan tabel menggunakan script SQL
mysql -h your-polardb-host -u username -p < setup_database.sql

# 2. Atau jalankan manual di MySQL client:
# CREATE DATABASE telegram_bot;
# USE telegram_bot;
# [jalankan isi setup_database.sql]
```

### 4. Jalankan dengan Docker Compose

```bash
# Build dan jalankan
docker-compose up --build -d

# Lihat logs
docker-compose logs -f qwen

# Stop bot
docker-compose down
```

### 5. Jalankan untuk Development

```bash
# Install dependencies
go mod tidy

# Jalankan bot
go run cmd/main.go
```

### 6. Testing WebSocket Interface

Setelah bot berjalan, Anda dapat mengakses interface web untuk testing streaming:

```
http://localhost:8080
```

Interface ini menyediakan:
- Real-time WebSocket connection
- Streaming response dengan thinking process
- Visual indicators untuk setiap tahap (thinking, reasoning, responding)
- Auto-reconnection jika koneksi terputus

## Struktur Project

```
Qwen/
├── cmd/
│   └── main.go              # Entry point aplikasi
├── internal/
│   ├── ai/
│   │   └── client.go        # Client AI dengan streaming support
│   ├── bot/
│   │   └── handler.go       # Handler Telegram bot dengan streaming
│   ├── config/
│   │   └── config.go        # Konfigurasi aplikasi
│   ├── server/
│   │   └── server.go        # HTTP server untuk WebSocket
│   └── websocket/
│       └── hub.go           # WebSocket hub dan client management
├── docker-compose.yml       # Docker compose configuration
├── Dockerfile              # Docker build instructions
├── env.example             # Template environment variables
├── go.mod                  # Go module dependencies
└── README.md              # Dokumentasi
```

## Penggunaan

### Telegram Bot
1. **Start Bot**: Kirim `/start` untuk memulai percakapan dengan AI yang personal
2. **Help**: Kirim `/help` untuk melihat bantuan
3. **Chat dengan AI**: Kirim pesan apapun untuk mengobrol dengan AI - akan melihat **streaming response real-time**:
   - ⚡ **Real-time streaming**: Respons muncul word-by-word dari API
   - 🔄 **Live updates**: Message terupdate secara incremental  
   - 💬 **Natural typing effect**: Seperti melihat AI mengetik secara langsung
   - 🚀 **Fast & reliable**: Optimized untuk performa tinggi

### WebSocket Interface
1. Buka `http://localhost:8080` di browser
2. Ketik pesan dan lihat **streaming response** real-time
3. Fitur interface:
   - **Real-time streaming**: Melihat respons muncul secara incremental
   - **Visual indicators**: Status connected/disconnected  
   - **Auto-reconnection**: Koneksi otomatis jika terputus
   - **Modern UI**: Interface yang clean dan responsive

## Command yang Tersedia

- `/start` - Memulai percakapan dengan bot
- `/help` - Menampilkan pesan bantuan
- `/resetmemory` - Menghapus semua memory/informasi personal yang tersimpan

## Fitur Memory System

Bot ini dilengkapi dengan sistem memory permanen yang dapat:

### 🧠 Deteksi Dinamis dengan LLM
- **Contextual Analysis**: LLM menganalisis konteks percakapan secara natural
- **No Regex Patterns**: Tidak perlu maintenance regex patterns manual
- **Flexible Information**: Dapat mendeteksi informasi kompleks dan beragam
- **Automatic Merging**: Menggabungkan informasi baru dengan memory lama
- **Smart Updates**: Memperbarui informasi yang sudah berubah secara otomatis

### 📝 Cara Kerja Memory LLM-Based
1. **Input Analysis**: LLM menganalisis pesan user + memory JSON lama
2. **Memory Processing**: AI memutuskan informasi mana yang penting dan perlu disimpan
3. **Dynamic Merge**: Informasi baru digabungkan dengan memory lama secara intelligent
4. **JSON Storage**: Memory disimpan dalam format JSON yang fleksibel
5. **Contextual Response**: AI memberikan respons personal berdasarkan memory terkini

### 💡 Contoh Penggunaan Dynamic Memory
```
User: Halo, nama saya Budi, umur 28, kerja sebagai programmer di Jakarta
Bot: Halo Budi! Senang berkenalan denganmu. Programmer di Jakarta pasti sibuk ya?

[Memory tersimpan: {"name": "Budi", "age": 28, "job": "programmer", "location": "Jakarta"}]

[Percakapan lanjutan...]
User: Sekarang aku lagi di Bali liburan, hobi aku fotografi
Bot: Wah Budi! Bali pasti indah banget untuk fotografi. Sebagai programmer yang hobi 
     fotografi, pasti banyak momen bagus yang bisa diabadikan di sana!

[Memory diupdate: location → "Bali", interests → ["fotografi"], dll]

[Kemudian...]
User: Rekomendasikan tempat makan dong
Bot: Halo Budi! Di Bali ada banyak kuliner enak nih. Karena kamu suka fotografi, 
     coba ke Ubud - ada restoran dengan view sawah yang instagramable banget!
```

### 🔒 Privacy & Control
- Memory bersifat personal per user (berdasarkan Telegram user ID)
- User dapat menghapus memory kapan saja dengan `/resetmemory`
- Jika database tidak tersedia, bot tetap berfungsi tanpa memory

## Contoh Actual Thinking Process

Ketika Anda bertanya: **"Bagaimana cara kerja blockchain?"**

Bot akan menampilkan:

### 🤔 Thinking (Real-time streaming):
```
Pertanyaan ini meminta penjelasan tentang blockchain... 
perlu menjelaskan konsep dasar... 
harus menyederhanakan teknologi kompleks...
mempertimbangkan audiens umum...
```

### 💭 Reasoning (Complete analysis):
```
💭 Alasan: Saya akan menjelaskan blockchain dengan:
1. Definisi sederhana terlebih dahulu
2. Analogi yang mudah dipahami
3. Komponen utama teknologi
4. Manfaat dan contoh penggunaan
```

### ✍️ Answer (Final response):
```
Blockchain adalah teknologi penyimpanan data digital yang...
[Jawaban lengkap berdasarkan analisis di atas]
```

**Semua tahap ini menggunakan streaming incremental dari Qwen API dengan `stream=true`**

## Konfigurasi

Bot dapat dikonfigurasi melalui environment variables:

- `TELEGRAM_BOT_TOKEN`: Token bot Telegram dari BotFather
- `DASHSCOPE_API_KEY`: API key dari Alibaba Cloud Model Studio
- `DASHSCOPE_BASE_URL`: Base URL untuk API (default: Singapore region)
- `AI_MODEL`: Model AI yang digunakan (default: qwen-mt-turbo)
- `HTTP_PORT`: Port untuk HTTP server dan WebSocket (default: 8080)

## Region API

Bot mendukung dua region:

- **Singapore**: `https://dashscope-intl.aliyuncs.com/compatible-mode/v1`
- **Beijing**: `https://dashscope.aliyuncs.com/compatible-mode/v1`

## Logging

Bot akan menampilkan log untuk:
- Pesan yang diterima dari pengguna
- Error saat berkomunikasi dengan API
- Status startup dan shutdown

## Troubleshooting

### Bot tidak merespons
- Pastikan `TELEGRAM_BOT_TOKEN` benar
- Periksa koneksi internet
- Lihat logs untuk error details

### Error AI response
- Pastikan `DASHSCOPE_API_KEY` valid
- Periksa `DASHSCOPE_BASE_URL` sesuai region
- Pastikan model `qwen-mt-turbo` tersedia

### Docker issues
- Pastikan Docker dan Docker Compose terinstall
- Periksa file `.env` sudah dibuat dan terisi
- Jalankan `docker-compose logs` untuk melihat error

## Development

Untuk development, pastikan Go 1.21+ terinstall:

```bash
# Install dependencies
go mod tidy

# Format code
go fmt ./...

# Run tests (jika ada)
go test ./...

# Build aplikasi
go build -o bot cmd/main.go
```

## Contributing

1. Fork repository
2. Buat feature branch
3. Commit perubahan
4. Push ke branch
5. Buat Pull Request

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0).

See the `LICENSE` file for details.
