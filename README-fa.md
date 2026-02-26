<div align="center">

```
████████╗ ██████╗ ██████╗
   ██╔══╝██╔═══██╗██╔══██╗
   ██║   ██║   ██║██████╔╝
   ██║   ██║   ██║██╔══██╗
   ██║   ╚██████╔╝██║  ██║
   ╚═╝    ╚═════╝ ╚═╝  ╚═╝
        M A N A G E R
```

**مجموعه مدیریت حرفه‌ای Tor — متن‌باز و رایگان**
نوشته‌شده کاملاً با Go — در دو قالب رابط ترمینال و اپلیکیشن دسکتاپ چندسکویی

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-violet?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows%20%7C%20macOS-lightgrey?style=flat-square)](https://github.com/yourname/tor-manager)
[![Wails](https://img.shields.io/badge/GUI-Wails%20v2-red?style=flat-square)](https://wails.io)
[![Svelte](https://img.shields.io/badge/Frontend-Svelte-FF3E00?style=flat-square&logo=svelte)](https://svelte.dev)

[English](README.md) · **فارسی**

</div>

---

## 🧅 tor-manager چیست؟

**tor-manager** یک مجموعه کامل و حرفه‌ای برای مدیریت Tor daemon محلی شماست — بدون نیاز به دست زدن به فایل‌های پیکربندی یا اجرای دستورات پیچیده.

این ابزار به شما دید لحظه‌ای از وضعیت مدار Tor می‌دهد، امکان تغییر Bridge و کشور خروجی را در چند ثانیه فراهم می‌کند، و پروتکل کنترل قدرتمند Tor را در قالب یک رابط کاربری زیبا و ارگونومیک ارائه می‌دهد — چه ترمینال را ترجیح دهید، چه پنجره دسکتاپ.

این ابزار **VPN نیست**. **افزونه مرورگر نیست**. یک لایه مدیریتی درجه اول است که روی نصب موجود Tor شما قرار می‌گیرد و تمام قابلیت‌های آن را از طریق یک رابط مناسب در اختیار شما می‌گذارد.

> **آخرین نسخه:** نسخه ۰.۱.۱ منتشر شد — رفع اشکالات import حلقوی و احراز هویت Controller. جزئیات را در [CHANGELOG](CHANGELOG.md) ببینید.

---

## ✨ ویژگی‌های کلیدی

### 🔌 اتصال و هویت
- **ردیابی Bootstrap لحظه‌ای** — پیشرفت راه‌اندازی Tor از ۰٪ تا ۱۰۰٪ با به‌روزرسانی زنده نمایش داده می‌شود
- **تأیید IP خروجی** — آدرس IP فعلی از طریق اندپوینت رسمی Tor Project (`check.torproject.org`) بررسی می‌شود تا مطمئن شوید واقعاً از یک Exit Node عبور می‌کنید
- **هویت جدید با یک کلیک** — دستور `SIGNAL NEWNYM` با Cooldown اجباری ۱۰ ثانیه‌ای (مطابق محدودیت Tor) اجرا می‌شود
- **نمایش موقعیت جغرافیایی** — کشور، شهر، ASN و Latency رفت‌وبرگشت Exit Node نمایش داده می‌شود
- **آدرس Proxy SOCKS5** — یک‌کلیکی `socks5://127.0.0.1:9050` را کپی کنید و هر اپلیکیشنی را از طریق Tor عبور دهید

### 🌉 مدیریت Bridge
- **پشتیبانی از تمام Transport** — obfs4، Snowflake، WebTunnel، meek-azure و اتصال مستقیم (Vanilla)
- **کتابخانه Bridge داخلی** — Bridge های معتبر و به‌روز برگرفته از Tor Browser برای هر نوع Transport
- **یکپارچگی با BridgeDB** — درخواست Bridge تازه مستقیماً از API رسمی MOAT پروژه Tor (`bridges.torproject.org`)، که خودِ درخواست نیز از طریق Tor مسیریابی می‌شود
- **ورود Bridge سفارشی** — هر خط Bridge معتبری را paste کنید؛ قبل از اعمال، parsing و اعتبارسنجی ساختاری انجام می‌شود
- **اعمال بدون Restart** — تغییرات Bridge از طریق Control Port روی Daemon در حال اجرا اعمال می‌شود

### 🌍 کنترل کشور خروجی
- **انتخاب چند کشور** — یک یا چند کشور به عنوان کاندیدای Exit Node انتخاب کنید
- **کلید StrictNodes** — فیلتر کشور را اجباری کنید (با هشدار کاهش ناشناسی)
- **اعمال فوری** — مقادیر `ExitNodes` و `StrictNodes` بلافاصله روی Daemon به‌روز می‌شود

### 📋 لاگ‌ها
- **جریان لاگ لحظه‌ای** — خروجی Tor Daemon به محض تولید نمایش داده می‌شود
- **رنگ‌بندی هوشمند** — رویدادهای Bootstrap، ساخت مدار، هشدارها و خطاها هرکدام استایل متمایز دارند
- **فیلتر سطح لاگ** — نمایش همه، Notice، Warn یا Error
- **جستجوی متن** — فیلتر خطوط لاگ بر اساس کلیدواژه
- **حالت Follow** — اسکرول خودکار به آخرین خط، یا قفل اسکرول برای مرور تاریخچه

### ⚙️ تنظیمات
- **پیکربندی پورت** — تغییر پورت SOCKS5 و Control بدون ویرایش دستی torrc
- **حالت‌های احراز هویت** — Cookie Auth (پیشنهادشده، امن‌تر) یا Password Auth
- **بازه Auto-Refresh** — تنظیم دوره بررسی خودکار IP خروجی
- **ذخیره پایدار** — تنظیمات در مسیر پیکربندی مناسب سیستم‌عامل ذخیره می‌شود (XDG در Linux، AppData در Windows، Library در macOS)

---

## 🏗️ معماری

این یک **Monorepo** است که از سه ماژول Go مستقل با یک کتابخانه Core مشترک تشکیل شده است.

```
tor-manager/
│
├── core/                        ← کتابخانه مشترک (بدون وابستگی UI)
│   ├── tor/
│   │   ├── controller.go        کلاینت Control Port (Thread-safe، Multi-auth)
│   │   ├── bridge.go            انواع Bridge، parsing، validation، اعمال زنده
│   │   ├── builtin_bridges.go   کتابخانه Bridge داخلی (obfs4/snowflake/webtunnel/meek)
│   │   ├── bridgedb.go          کلاینت BridgeDB MOAT API (محدودیت اندازه، مسیریابی از Tor)
│   │   ├── config.go            تولید torrc با نوشتن اتمیک (mode 0600)
│   │   ├── process.go           چرخه حیات Tor Daemon + ردیابی Bootstrap
│   │   ├── cookie.go            خواننده فایل Cookie چندسکویی
│   │   └── errors.go            انواع خطای Sentinel
│   ├── proxy/
│   │   └── checker.go           تأیید IP از طریق check.torproject.org + ipinfo.io
│   └── config/
│       └── appconfig.go         تنظیمات کاربر (XDG-aware، ذخیره اتمیک)
│
├── tor-tui/                     ← رابط ترمینال (Bubble Tea + Lip Gloss)
│   ├── main.go
│   └── ui/
│       ├── common/              ← انواع مشترک (بدون وابستگی UI)
│       │   ├── messages.go      انواع پیام Tea برای ارتباط بین کامپوننت‌ها
│       │   ├── styles.go        سیستم طراحی مرکزی (پالت رنگ، کامپوننت‌ها، توابع کمکی)
│       │   └── keys.go          تمام Keybinding ها در یک فایل
│       ├── app.go               مدل ریشه: مسیریابی Tab، صفحه Boot، Help Overlay
│       └── views/
│           ├── dashboard.go     وضعیت اتصال، IP خروجی، آدرس Proxy
│           ├── bridges.go       مدیریت Bridge (داخلی / سفارشی / BridgeDB)
│           ├── countries.go     انتخاب کشور خروجی با Multi-select
│           ├── logs.go          نمایشگر لاگ اسکرول‌پذیر با Follow Mode
│           └── settings.go      ویرایشگر تنظیمات با validation زنده
│
└── tor-gui/                     ← دسکتاپ GUI (Wails v2 + Svelte + Tailwind)
    ├── main.go                  پیکربندی پنجره
    ├── app.go                   Backend Go — تمام متدها از طریق IPC در دسترس Svelte
    └── frontend/
        └── src/
            ├── App.svelte       ریشه: نوار عنوان، Sidebar Nav، اتصال رویدادها
            ├── app.css          سیستم طراحی (Tailwind + کامپوننت‌های سفارشی)
            ├── lib/
            │   ├── wails.js     Bridge IPC ایمن (با Mock برای حالت dev)
            │   └── stores.js    State واکنشی Svelte (connected، ipInfo، logs، ...)
            └── components/
                ├── Dashboard.svelte
                ├── Bridges.svelte
                ├── Countries.svelte
                ├── Logs.svelte
                └── Settings.svelte
```

---

## 🔒 طراحی امنیتی

امنیت در این پروژه یک الزام اولیه است، نه یک فکر ثانویه.

| لایه | اقدام امنیتی |
|------|-------------|
| **احراز هویت Control Port** | SafeCookie با HMAC Challenge-Response پشتیبانی می‌شود؛ رمز عبور برای جلوگیری از Injection sanitize می‌شود |
| **GETINFO / SETCONF** | کلیدها در برابر یک regex دقیق Allowlist می‌شوند — هیچ رشته دلخواهی نمی‌تواند به Control Port برسد |
| **تولید torrc** | نوشتن اتمیک از طریق Temp-file + Rename؛ مجوز فایل `0600` (فقط owner می‌تواند بخواند) |
| **درخواست‌های BridgeDB** | بدنه پاسخ به ۶۴ کیلوبایت محدود است تا از تخلیه حافظه جلوگیری شود؛ تا جای ممکن از طریق Tor مسیریابی می‌شود |
| **SafeSocks** | به صورت پیش‌فرض فعال (`SafeSocks 1`) — درخواست‌های SOCKS4 که DNS leak دارند رد می‌شوند |
| **Stream Isolation** | `IsolateClientAddr 1` و `IsolateClientProtocol 1` به صورت پیش‌فرض فعال هستند |
| **HTTP Client** | `DisableKeepAlives` روی Proxy Checker — هر درخواست یک اتصال TCP تازه می‌گیرد |
| **Cooldown NEWNYM** | ۱۰ ثانیه حداقل، در کد اجباری شده — از تغییر سریع هویت تصادفی جلوگیری می‌کند |
| **فایل پیکربندی** | با الگوی Temp+Rename ذخیره می‌شود؛ هرگز نوشتن ناقص رخ نمی‌دهد |
| **مسیر Cookie** | قبل از خواندن در برابر Path Traversal اعتبارسنجی می‌شود |

---

## 🚀 شروع سریع

### پیش‌نیازها

```bash
# Tor Daemon
sudo apt install tor                  # Debian / Ubuntu
brew install tor                      # macOS
# ویندوز: دانلود از https://www.torproject.org

# Transport های قابل اتصال (برای پشتیبانی از Bridge)
sudo apt install obfs4proxy           # obfs4 + meek-azure

# فقط برای tor-gui
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# وابستگی‌های GUI در Linux
sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev
```

**فعال‌سازی Control Port** در `torrc`:

```
# /etc/tor/torrc  (Linux)
# %APPDATA%\Tor\torrc  (Windows)
# ~/Library/Application Support/TorBrowser-Data/Tor/torrc  (macOS)

ControlPort 9051
CookieAuthentication 1
```

سپس Tor را Restart کنید:

```bash
sudo systemctl restart tor       # Linux
brew services restart tor        # macOS
# ویندوز: از طریق Services یا Tor Browser
```

---

### Clone و Build

```bash
git clone https://github.com/amirk1998/tor-manager.git
cd tor-manager
```

#### رابط ترمینال (tor-tui)

```bash
cd tor-tui

# دریافت وابستگی‌ها
go mod download

# اجرای مستقیم
go run .

# Build باینری
go build -ldflags "-s -w" -o bin/tor-tui .

# ویندوز
go build -ldflags "-s -w" -o bin/tor-tui.exe .
```

#### دسکتاپ GUI (tor-gui)

```bash
cd tor-gui

# نصب وابستگی‌های فرانت‌اند
cd frontend && npm install && cd ..

# حالت توسعه — Hot Reload برای هم Go و هم Svelte
wails dev

# Build نهایی
wails build -clean -ldflags "-s -w"

# با فشرده‌سازی UPX (~۴۰٪ کوچکتر)
wails build -clean -upx

# Cross-compile برای ویندوز (از Linux/macOS)
wails build -platform windows/amd64
```

---

## 🖥️ راهنمای استفاده

### کلیدهای TUI

| کلید | عملکرد |
|------|--------|
| `1` تا `5` | جابجایی بین تب‌ها |
| `tab` / `shift+tab` | تب بعدی / قبلی |
| `?` | نمایش/پنهان‌کردن راهنمای کلیدها |
| `q` / `ctrl+c` | خروج |
| **Dashboard** | |
| `r` | به‌روزرسانی IP خروجی |
| `n` | هویت جدید (Cooldown ۱۰ ثانیه‌ای) |
| `c` | کپی آدرس Proxy |
| **Bridges** | |
| `←` `→` | تغییر نوع Transport |
| `↑` `↓` / `enter` | ناوبری و انتخاب Bridge |
| `u` | افزودن Bridge سفارشی |
| `f` | دریافت Bridge جدید از BridgeDB |
| `a` | اعمال پیکربندی Bridge فعلی |
| `x` | غیرفعال‌کردن همه Bridge ها |
| **Countries** | |
| `↑` `↓` | ناوبری در لیست کشورها |
| `enter` / `space` | Toggle انتخاب کشور |
| `s` | Toggle کردن StrictNodes |
| `a` | اعمال فیلتر کشور |
| `x` | پاک‌کردن فیلتر (هر کشوری) |
| **Logs** | |
| `↑` `↓` | اسکرول |
| `g` / `G` | رفتن به ابتدا / انتها |
| `f` | Toggle حالت Follow |
| `c` | پاک‌کردن بافر لاگ |

### اتصال اپلیکیشن‌ها به Tor

پس از اجرا، تنظیمات Proxy هر اپلیکیشنی را وارد کنید:

```
Protocol:  SOCKS5
Host:      127.0.0.1
Port:      9050
```

**مثال‌ها:**

```bash
# تلگرام: Settings → Privacy and Security → Use proxy → SOCKS5, Host: 127.0.0.1, Port: 9050

# curl
curl --socks5 127.0.0.1:9050 https://check.torproject.org/api/ip

# git
git config --global http.proxy socks5://127.0.0.1:9050

# متغیر محیطی (بسیاری از ابزارهای CLI)
export ALL_PROXY=socks5://127.0.0.1:9050

# Python requests
import requests
proxies = {'http': 'socks5://127.0.0.1:9050', 'https': 'socks5://127.0.0.1:9050'}
r = requests.get('https://check.torproject.org/api/ip', proxies=proxies)
```

---

## 🔧 Stack فناوری

| بخش | فناوری | دلیل انتخاب |
|-----|--------|------------|
| کتابخانه Core | Go 1.22 | عملکرد بالا، Binary واحد، بدون CGO |
| فریمورک TUI | [Bubble Tea](https://github.com/charmbracelet/bubbletea) | معماری Elm — State قابل پیش‌بینی |
| استایل TUI | [Lip Gloss](https://github.com/charmbracelet/lipgloss) | Layout ترمینال declarative |
| فریمورک GUI | [Wails v2](https://wails.io) | Go + Web UI، بدون CGO، بدون Electron |
| فرانت‌اند GUI | [Svelte 4](https://svelte.dev) | بدون Virtual DOM، کوچکترین Bundle |
| استایل GUI | [Tailwind CSS 3](https://tailwindcss.com) | Utility-first، Purge در Production |
| Build Tool | [Vite](https://vitejs.dev) | HMR سریع، Build بهینه Production |
| فشرده‌سازی | [UPX](https://upx.github.io) | کوچک‌کردن اختیاری Binary |

---

## 🛠️ توسعه

### ساختار ماژول‌ها

هر ماژول `go.mod` مستقل خود را دارد و در حین توسعه از directive جایگزینی استفاده می‌کند:

```go
// tor-tui/go.mod  و  tor-gui/go.mod
replace github.com/amirk1998/tor-manager/core => ../core
```

### اجرای تست‌ها

```bash
# تست کتابخانه Core
cd core && go test ./... -v

# با Race Detector
cd core && go test -race ./...
```

### دستورات سراسری (Makefile ریشه)

```bash
make dev-tui        # اجرای TUI
make dev-gui        # اجرای GUI با Hot Reload
make build-tui      # Build باینری TUI
make build-gui      # Build باینری GUI
make build-all      # Build هر دو
make tidy           # go mod tidy برای همه ماژول‌ها
make test           # اجرای همه تست‌ها
```

---

## 📦 فایل‌های Release

| سکو | TUI | GUI |
|-----|-----|-----|
| Linux x64 | `tor-tui-linux-amd64` | `tor-gui` |
| Windows x64 | `tor-tui-windows.exe` | `tor-gui.exe` |
| macOS x64 | `tor-tui-macos-amd64` | `tor-gui.app` |
| macOS ARM64 | `tor-tui-macos-arm64` | `tor-gui-arm64.app` |

حجم تقریبی باینری‌ها (با UPX):

| باینری | بدون UPX | با UPX |
|--------|---------|-------|
| tor-tui | ~۸ مگابایت | ~۳ مگابایت |
| tor-gui | ~۱۵ مگابایت | ~۸ مگابایت |

---

## 🗺️ نقشه راه

- [ ] **v0.2** — Daemon Tor تعبیه‌شده (بدون نیاز به نصب جداگانه)
- [ ] **v0.2** — Installer ویندوز (NSIS) از طریق `wails build -nsis`
- [ ] **v0.3** — نمایشگر مدار Tor (نمایش Relay ها روی نقشه جهانی)
- [ ] **v0.3** — مدیریت Onion Service (ایجاد و مدیریت آدرس‌های `.onion`)
- [ ] **v0.4** — به‌روزرسانی خودکار لیست Bridge داخلی از CDN پروژه Tor
- [ ] **v0.4** — یکپارچگی با System Tray برای GUI

---

## 🤝 مشارکت

مشارکت‌ها خوش‌آمد است. لطفاً از قرارداد Commit پیروی کنید:

```
<type>(<scope>): <subject>

type:    feat | fix | chore | refactor | docs | test
scope:   core | tui | gui | bridges | countries | logs

مثال‌ها:
  feat(core): add SafeCookie authentication support
  fix(tui): resolve NEWNYM cooldown timer reset bug
  feat(gui): implement BridgeDB fetch with spinner
  chore(deps): update bubbletea to v0.26.4
```

**استراتژی Branch:**
- `main` — انتشارهای پایدار و Tag‌گذاری‌شده
- `dev` — Branch یکپارچه‌سازی، پایه همه Branch های Feature
- `feat/*` — ویژگی‌های جداگانه، PR به `dev`

---

## 📄 مجوز

مجوز MIT — متن کامل را در [LICENSE](LICENSE) ببینید.

---

## ⚠️ سلب مسئولیت

این نرم‌افزار با شبکه Tor ارتباط برقرار می‌کند. استفاده از Tor در اکثر کشورها قانونی است، اما شما مسئول درک و رعایت قوانین حوزه قضایی خود هستید. این ابزار به تنهایی ناشناسی فراهم نمی‌کند — این ابزار Tor Daemon شما را مدیریت می‌کند. همیشه از Tor به شکل مسئولانه استفاده کنید.

---

<div align="center">

ساخته‌شده با Go · Bubble Tea · Wails · Svelte · Tailwind

**اگر این پروژه برای شما مفید است، در GitHub ستاره بدهید ⭐**

</div>
