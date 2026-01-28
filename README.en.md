<div align="center">
  <h1>🌟 Custom Branch Features</h1>

  <p>
    <a href="#new-features">✨ New Features</a> • 
    <a href="#migration-guide">🔄 Migration Guide</a> • 
    <a href="#custom-version-deployment">🐳 Deployment</a> •
    <a href="#ech0">📖 Original Info</a>
  </p>
</div>

> This branch (`custom`) is an enhanced version based on the original [`main`](https://github.com/lin-snow/Ech0) branch, with many new features and optimizations.
>
> 💡 **Special Notice**
>
> This is a **personal customized version**. While maintaining synchronization with the original core, it includes features that may deviate from the original "minimalistic" design philosophy.
>
> **Review your needs before choosing:**
> 1. **Original Recommended**: Use the [original version](https://github.com/lin-snow/Ech0) for the purest Ech0 experience.
> 2. **Use as Needed**: Choose this version only if you require specific features like Video or Live Photo support.
> 3. **Seamless Switch**: We provide measures for [Bi-directional Migration](#migration-guide), allowing you to switch back at any time.
> 4. **Sync Policy**: We track upstream updates promptly. In case of functional conflicts, upstream code takes precedence.
> 5. **Performance Note**: To ensure compatibility and ease of syncing, some extra logic (e.g., real-time legacy table sync) is implemented, which may cause negligible performance overhead. Please consider this if you have extreme performance requirements.

### 🐳 Quick Deployment

```shell
docker run -d --name ech0 -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -v /opt/ech0/backup:/app/backup \
  -e JWT_SECRET="your-secret-key" \
  lumlime/ech0:latest
```

> Access `http://ip:6277` after deployment | First registered user becomes admin | [Detailed deployment guide](#-custom-version-deployment)

</div>

<details open>
<summary><strong>📋 New Features</strong></summary>

### 🎬 Media System Refactoring
- **Video Upload Support** — Media storage system refactored from `images` to `media`, supporting mixed image and video uploads
- **Video Configuration** — Backend adds video upload config (max 100MB) and storage path settings
- **Live Photo Support** — Full iOS/Android Live Photo support:
  - Auto-detect embedded motion photos and separate into image + video
  - Auto-pair images and videos with matching filenames
  - Auto-play live photos in Fancybox preview (configurable auto-play toggle)
  - Mobile-optimized live photo viewing experience
  - Dedicated live photo icon and LIVE badge
- **HEIC/HEIF Format Support** — Auto-convert to JPEG format on upload
- **Media Drag & Drop Sorting** — Drag to reorder media items, mobile touch threshold to prevent accidental drags
- **Legacy Data Compatibility** — Auto-migrate `images` to `media` data format
- **Unsupported Media Type Hint** — Shows friendly placeholder when legacy clients access Echos with video

### 🤖 AI Smart Layout Recommendation
- **Auto Layout Recommendation** — AI/rule engine intelligently recommends optimal image layout based on media info and content analysis
- **Deep Content Analysis** — Analyzes text structure (code blocks, links, headers, lists, etc.) to optimize recommendations
- **Text Semantic Analysis** — AI understands user intent (expressing views, showcasing works, documenting journeys, teaching/comparing)
- **Recommendation Reasons** — Displays recommendation source (AI/Rule) and specific reasons
- **New Auto Layout Mode** — Added "Auto" option in layout selector, uses AI recommendation by default

### 📅 Calendar Heatmap
- **Calendar View Mode** — Heatmap supports switching to calendar view, browse by year/month
- **Date Filtering** — Click on calendar dates to filter Echos for that day
- **Year/Month Filtering** — Click year/month title to filter entire month's Echos
- **Filter Label Display** — Top navigation shows current filter date/month with click-to-cancel
- **Mobile Gestures** — Long-press to switch view modes, tap to filter with touch interactions
- **Year/Month Switcher** — Quickly browse historical monthly publishing data
- **View Mode Memory** — Remembers user's selected heatmap view mode

### ✅ Todo Enhancements
- **Widget Completion** — Complete todos directly from the todo widget card
- **Undo Completion** — Undo recently completed todos
- **Todo Blinking Reminder** — Widget icon blinks when there are pending todos
- **New Checkbox Component** — Newly designed checkbox component with animations

### 🔔 Hub Update Notifications
- **Update Badge Display** — Shows red badge with update count when Hub has new content
- **Per-Site Statistics** — Displays update count for each subscribed site
- **Tooltip Details** — Hover to show per-site update counts
- **Mobile Bubble Hint** — Touch to display update details bubble
- **Background Polling** — Auto-detect updates without manual refresh
- **Window Focus Refresh** — Auto-check updates when switching back to page

### ⏰ Time Display Optimization
- **Click to Switch Format** — Click to toggle time display format (relative/absolute)
- **Smart Time Display** — Auto-select best display format based on time distance

### ✏️ Editor Enhancements
- **Draft Auto-Save** — Editor content auto-saved locally to prevent accidental loss
- **Draft Recovery** — Recover drafts after page refresh or accidental close
- **Update Mode Detection** — Smart detection of actual changes when editing existing Echos
- **Empty Draft Auto-Cleanup** — Auto-cleanup local drafts when content is empty
- **Editor Save Status** — Editor toolbar shows draft save status and timestamp
- **Grid Media Preview** — Editor media preview uses 3x3 grid layout
- **Editor User Avatar** — Editor title bar shows user avatar and username when logged in

### 🔐 OAuth Login Optimization
- **QQ Login Re-enabled** — Refactored OAuth2 login flow, re-enabled QQ Connect login
- **Registration Permission Check** — Auto-check system registration permissions during OAuth login
- **Dynamic Register Button** — Show/hide register button based on system settings

### 🎨 Site & User Configuration
- **Independent Site Logo** — Site logo separated from user avatar, supports individual configuration
- **Echo Shows User Avatar** — Echo detail page shows publisher's avatar instead of site logo
- **Theme Auto Mode** — Theme follows system settings with current mode status display

### 🏗️ CI/CD & Deployment
- **Docker Image Auto-Build** — Auto-build and push Docker images on push to custom branch
- **Dockerfile Optimization** — Simplified build process with unified build (frontend + backend together)
- **MIME Type Support** — Added mailcap to Docker image for extended MIME type recognition

### 📱 PWA & SEO Enhancements
- **PWA Info Auto-Sync** — Automatically syncs site title, description, and logo to PWA config for a unified installation experience
- **Smart Logo Cropping** — Automatically performs center-square cropping and resizing upon logo upload for perfect icon display
- **Dynamic Social Sharing Content** — Deeply optimized OpenGraph support; sharing links on platforms like X (Twitter), Telegram, or Discord now automatically generates rich preview cards with custom text summaries and media
- **Smart Page Titles** — Title bar automatically reflects the current context, such as Echo details, specific search terms, or date filters, making tab navigation effortless

### 🔧 Other Optimizations
- **RSS Media Display** — Optimized RSS attachment display based on media type, distinguishing video and image
- **ActivityPub Video Support** — Fediverse attachment types auto-distinguished as Image/Video/Document
- **Video Thumbnail Optimization** — Fixed video thumbnail display issues in some browsers
- **Tag Query Fix** — Fixed tag association query condition errors
- **Auto-Scroll After Edit** — Auto-scroll to corresponding position after updating Echo
- **Filter List Sync Update** — Filter list auto-syncs after editing Echo
- **Auto-Delete Empty Echo** — Auto-delete Echo when all media is removed and content is empty
- **Atomic Live Photo Deletion** — Deleting live photo also removes associated video
- **In-App Browser Compatibility** — Optimized link navigation in WeChat and other in-app browsers
- **Hub Data Caching** — Optimized Hub data requests with caching mechanism to reduce duplicate requests

</details>

<details>
<summary><strong>📦 Data Structure Changes</strong></summary>

### API Changes
- `images` field renamed to `media`
- Added `media_type` field (`image` / `video`)
- Added `live_video_id` field for live photo associations
- Added `live_pair_id` field for live photo pairing during upload
- `image_url` / `image_source` renamed to `media_url` / `media_source`
- Echo response includes new `user` field with publisher info

### Database Migration
- `images` table auto-migrated to `media` table
- Migration auto-sets `media_type` to `image` for all existing images

### Configuration Changes
- Added `videomaxsize` config (max video upload size)
- Added `videopath` config (video storage path)
- `allowedtypes` extended with video formats (mp4, webm, quicktime)
- `allowedtypes` extended with HEIC/HEIF formats

### Backward Compatibility
- Frontend auto-handles `images` field from legacy servers
- Backend Hub connection auto-converts legacy data format
- Legacy clients see friendly hint when accessing Echos with video

</details>

<details>
<summary><strong>🐳 Custom Version Deployment</strong></summary>

### Docker Deployment

Custom branch uses an independent Docker image repository:

```shell
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -v /opt/ech0/backup:/app/backup \
  -e JWT_SECRET="Hello Echos" \
  lumlime/ech0:latest
```

> 💡 After deployment, access `ip:6277` to use  
> 🚷 It is recommended to change `JWT_SECRET="Hello Echos"` to a secure secret  
> 📍 The first registered user will be set as administrator  
> 🎈 Data stored under `/opt/ech0/data`

### Upgrading

```shell
# Stop the current container
docker stop ech0

# Remove the container
docker rm ech0

# Pull the latest image
docker pull lumlime/ech0:latest

# Start the new version
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -v /opt/ech0/backup:/app/backup \
  -e JWT_SECRET="Hello Echos" \
  lumlime/ech0:latest
```

</details>

---

<p align="left">
  <a href="https://hellogithub.com/repository/lin-snow/Ech0" target="_blank">
    <img src="https://api.hellogithub.com/v1/widgets/recommend.svg?rid=8f3cafdd6ef3445dbb1c0ed6dd34c8b5&claim_uid=swhbQfnJvKS0t7I&theme=neutral"
         alt="Featured｜HelloGitHub"
         width="250"
         height="54" />
  </a>
</p>

<p align="right">
  <a title="zh" href="./README.md">
    <img src="https://img.shields.io/badge/-简体中文-545759?style=for-the-badge" alt="简体中文">
  </a>
  <img src="https://img.shields.io/badge/-English-F54A00?style=for-the-badge" alt="English">
</p>



<div align="center">
  <img alt="Ech0" src="./docs/imgs/logo.svg" width="150">

  [Preview](https://memo.vaaat.com/) | [Official Site & Doc](https://www.ech0.app/) | [Ech0 Hub](https://hub.ech0.app/)

  # Ech0
</div>

<div align="center">

[![GitHub release](https://img.shields.io/github/v/release/lin-snow/Ech0)](https://github.com/lin-snow/Ech0/releases) ![License](https://img.shields.io/github/license/lin-snow/Ech0) [![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/lin-snow/Ech0) [![Hello Github](https://api.hellogithub.com/v1/widgets/recommend.svg?rid=8f3cafdd6ef3445dbb1c0ed6dd34c8b5&claim_uid=swhbQfnJvKS0t7I&theme=small)](https://hellogithub.com/repository/lin-snow/Ech0)

</div>

> A next-generation open-source, self-hosted, lightweight federated publishing platform focused on personal idea sharing.

Ech0 is a new-generation open-source self-hosted platform designed for individual users. It is ultra-lightweight and low-cost, supporting the ActivityPub protocol to let you easily publish and share ideas, writings, and links. With a clean, intuitive interface and powerful command-line tools, content management becomes simple and flexible. Your data is fully owned and controlled by you, always connected to the world, building your own network of thoughts.

![Interface Preview](./docs/imgs/screenshot.png)

---

<details>
   <summary><strong>Table of Contents</strong></summary>

- [Ech0](#ech0)
  - [Highlights](#highlights)
  - [Quick Deployment](#quick-deployment)
    - [🐳 Docker (Recommended)](#-docker-recommended)
    - [🐋 Docker Compose](#-docker-compose)
    - [☸️ Kubernetes (Helm)](#️-kubernetes-helm)
  - [Upgrading](#upgrading)
    - [🔄 Docker](#-docker)
    - [💎 Docker Compose](#-docker-compose-1)
    - [☸️ Kubernetes (Helm)](#️-kubernetes-helm-1)
  - [FAQ](#faq)
  - [Feedback \& Community](#feedback--community)
  - [Architecture](#architecture)
  - [Development Guide](#development-guide)
    - [Backend Requirements](#backend-requirements)
    - [Frontend Requirements](#frontend-requirements)
    - [Start Backend \& Frontend](#start-backend--frontend)
  - [Thanks for Your Support!](#thanks-for-your-support)
  - [Star History](#star-history)
  - [Acknowledgements](#acknowledgements)
  - [Support](#support)
</details>

---

<div id="migration-guide"></div>
<details open>
<summary><strong>🔄 Branch Migration & Original Compatibility</strong></summary>

### 1. Migrating from Original (`main`) to this branch (`custom`)
- **Auto Migration**: Simply deploy this branch's image and mount your existing database files.
- **How it works**: Upon startup, the program automatically detects records in the legacy `images` table and **incrementally syncs** them to the new `media` table.
- **Lossless Upgrade**: All your original images, configurations, and Echos will be preserved intact.
- **Delete Sync**: When you delete an image in this branch, it will also be removed from the legacy `images` table to ensure the deletion takes effect if you switch back.
  - ⚠️ **Note**: This auto-sync is **one-way**. Deletions made in the original version **will not** be automatically synced back to this branch.
  - 💡 **Suggestion**: Once you've decided on a version, avoid frequent switching. The current policy prioritizes data safety for this branch, thus "delete" operations from the original version are not back-synced.

### 2. Rolling back from this branch (`custom`) to Original (`main`)
Since this branch introduces new structures like video support, follow these steps to roll back:
1. **Sync Database**: Go to `Admin Panel -> Storage -> Original Compatibility` and click the **"Rebuild Data"** icon.
2. **Sync Result**: The system will write image records from the `media` table back to the legacy `images` table using ID alignment.
3. **Switch Version**: Once synced, stop the current container and switch back to the `main` branch image.
4. **⚠️ Limitations**: As the original version does not support videos, **Video and Live Photo** content will not be displayed after rolling back.

### 3. Cleanup Space
If you decide to stick with the `custom` branch and have no plans to roll back, you can click **"Cleanup Data"** on the same settings page to delete legacy tables and free up database space.

</details>

---

## Highlights

☁️ **Atomically Lightweight**: Consumes less than **15MB** of memory with an image size under **50MB**, powered by a single-file SQLite architecture  
🚀 **Instant Deployment**: Zero configuration required — from installation to operation in just one command  
✍️ **Distraction-Free Writing**: A clean, online Markdown editor with rich plugin support and real-time preview  
📦 **Data Sovereignty**: All content is stored locally in SQLite, with full RSS feed support  
🔐 **Secure Backup Mechanism**: One-click export and full data backup across Web, TUI, and CLI modes, with automatic background backup support  
♻️ **Seamless Recovery**: Supports TUI/CLI snapshot restoration and Web-based zero-downtime recovery, ensuring data safety with ease  
🎉 **Forever Free**: Open-sourced under the AGPL-3.0 license — no tracking, no subscriptions, no external dependencies  
🌍 **Cross-Platform Adaptation**: Fully responsive design optimized for desktop, tablet, and mobile browsers  
👾 **PWA Ready**: Installable as a web application, offering a near-native experience  
🏷️ **Elegant Tag Management & Filtering**: Intelligent tagging system with fast filtering and precise search for effortless organization  
☁️ **S3 Storage Integration** — Native support for S3-compatible object storage enables efficient cloud synchronization  
🌐 **ActivityPub Federation** — Seamlessly federates with Mastodon, Misskey, and other decentralized platforms  
🔑 **OAuth2 & OIDC Authentication** — Native support for OAuth2 and OIDC protocols, enabling seamless third-party login and API authorization  
🙈 **Passkey Passwordless Login**: Supports passkey login based on biometrics or hardware keys, greatly enhancing security and login experience  
🪶 **Highly Available Webhook**: Enables real-time integration and collaboration with external systems, supporting event-driven automated workflows  
📝 **Built-in Todo Management**: Easily capture and manage daily tasks to stay organized and productive  
🧘 **Quiet Inbox Mode**: Minimizes system-level interruptions by default—messages are surfaced only as needed, letting the tool assist without intruding.
🌗 **Dark Mode & Theme Extensions**: Supports adaptive system dark mode or manual switching, with future extensibility for custom color schemes  
🤖 **Quick Agent AI Setup**: Easily configure multiple large language models for instant AI experience, no manual setup required  
🧰 **Command-Line Powerhouse**: A built-in high-availability CLI that empowers developers and advanced users with precision control and seamless automation  
🔑 **Quick Access Token Management**: Generate and revoke access tokens with one click for secure and efficient API calls and third-party integrations  
📊 **Real-Time System Resource Monitoring**: High-performance WebSocket-based monitoring dashboard for instant visibility into runtime status  
📟 **Refined TUI Experience**: A beautifully designed terminal interface offering intuitive management of Ech0  
🔗 **Ech0 Connect**: A multi-instance connectivity feature that enables real-time status sharing and synchronization between Ech0 nodes  
🎵 **Seamless Music Integration**: Lightweight embedded music player providing immersive soundscapes and focus modes  
🎥 **Instant Video Sharing**: Natively supports intelligent parsing of Bilibili and YouTube videos  
🃏 **Rich Smart Cards**: Instantly share websites, GitHub projects, and other media in visually engaging cards  
⚙️ **Advanced Customization**: Easily personalize styles and scripts for expressive, unique content presentation  
💬 **Comment System**: Quick Twikoo integration for lightweight, instant, and non-intrusive interactions  
💻 **Cross-Platform Compatibility**: Runs natively on Windows, Linux, and ARM devices like Raspberry Pi for stable deployment anywhere  
🔗 **Ech0 Hub Square**: Built-in Ech0 Hub Square for easily discovering, subscribing to, and sharing high-quality content  
📦 **Self-Contained Binary**: Includes all required resources — no extra dependencies, no setup hassle  
🔗 **Rich API Support**: Open APIs for seamless integration with external systems and workflows  
🃏 **Dynamic Content Display**: Supports Twitter-like card layouts with likes and social interactions  
👤 **Multi-Account & Permission Management**: Flexible user and role-based access control ensuring privacy and security  


---

## Quick Deployment
<!-- 
### 🧙 One-Click Script Deployment (Recommended, make sure your network can access GitHub Release)
```shell
curl -fsSL "https://sh.soopy.cn/ech0.sh" -o ech0.sh && bash ech0.sh
``` -->

### 🐳 Docker (Recommended)

```shell
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -v /opt/ech0/backup:/app/backup \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

> 💡 After deployment, access `ip:6277` to use  
> 🚷 It is recommended to change `JWT_SECRET="Hello Echos"` to a secure secret  
> 📍 The first registered user will be set as administrator  
> 🎈 Data stored under `/opt/ech0/data`

### 🐋 Docker Compose

1. Create a new directory and place `docker-compose.yml` inside.  
2. Run:

```shell
docker-compose up -d
```

### ☸️ Kubernetes (Helm)

If you want to deploy Ech0 in a Kubernetes cluster, you can use the Helm Chart provided in this project.

Since this project does not provide an online Helm repository, you need to clone the repository to your local machine first, and then install from the local directory.

1.  **Clone the repository:**
    ```shell
    git clone https://github.com/lin-snow/Ech0.git
    cd Ech0
    ```

2.  **Install with Helm:**
    ```shell
    # helm install <release-name> <chart-directory>
    helm install ech0 ./charts/ech0
    ```

    You can also customize the release name and namespace:
    ```shell
    helm install my-ech0 ./charts/ech0 --namespace my-namespace --create-namespace
    ```

---

## Upgrading

### 🔄 Docker

```shell
docker stop ech0
docker rm ech0
docker pull sn0wl1n/ech0:latest
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -v /opt/ech0/backup:/app/backup \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

### 💎 Docker Compose

```shell
cd /path/to/compose
docker-compose pull && \
docker-compose up -d --force-recreate
docker image prune -f
```

### ☸️ Kubernetes (Helm)

1.  **Update the repository:**
    Navigate to your local Ech0 repository directory and pull the latest changes.
    ```shell
    cd Ech0
    git pull
    ```

2.  **Upgrade the Helm Release:**
    Use the `helm upgrade` command to update your release.
    ```shell
    # helm upgrade <release-name> <chart-directory>
    helm upgrade ech0 ./charts/ech0
    ```
    If you used a custom release name and namespace, use the corresponding names:
    ```shell
    helm upgrade my-ech0 ./charts/ech0 --namespace my-namespace
    ```

<!-- ---

## Access Modes

### 🖥️ TUI Mode

![TUI Mode](./docs/imgs/tui.png)

Run the binary directly (for example, on Windows double-click `Ech0.exe`).

### 🔐 SSH Mode

Connect to the instance via port 6278:

```shell
ssh -p 6278 ssh.vaaat.com
``` -->

---

## FAQ

1. **What is Ech0?**  
   A lightweight, open-source self-hosted platform for quickly sharing thoughts, writings, and links. All content is locally stored.  

2. **What Ech0 is NOT?**  
   Not a professional note-taking app like Obsidian or Notion; its core function is similar to social feed/microblog.  

3. **Is Ech0 free?**  
   Yes, fully free and open-source under AGPL-3.0, no ads, tracking, subscription, or service dependency.  

4. **How do I back up and restore data?**  
  Since all content is stored in a local SQLite file, you only need to back up the files in the `/opt/ech0/data` directory (or the mapped path you chose during deployment). To restore, simply replace the data files with your backup. You can also use the online data management features in the settings under "Data Management" to quickly create, export, or restore snapshots. If the latest content does not appear after restoring, try manually restarting the Docker container.

5. **Does Ech0 support RSS?**  
   Yes, content updates can be subscribed via RSS.  

6. **Why can't I publish content?**  
   Only administrators can publish. First registered user is admin.  

7. **Why no detailed permission system?**  
   Ech0 emphasizes simplicity: admin vs non-admin only, for smooth experience.  

8. **Why Connect avatars may not show?**  
   Set your instance URL in `System Settings - Service URL` (with `http://` or `https://`).  

9. **What is MetingAPI?**  
   Used to parse music streaming URLs for music cards. If empty, default API provided by Ech0 is used.  

10. **Why not all Connect items show?**  
    Instances that are offline or unreachable are ignored; only valid instances are displayed.  

11. **What content is not recommended?**  
    Avoid publishing dense content mixing text + images + extension cards. Long posts or extension cards alone are okay.  

12. **How to enable comments?**  
    Set up Twikoo backend URL in settings. Only Twikoo is supported.  

13. **How to configure S3?**  
    Fill in endpoint (without http/https) and bucket with public access.

14. **How to join the Fediverse?**  
  You need to bind Ech0 to a domain name and fill in the domain in the server address field in the settings page. Once set, Ech0 will automatically join the Fediverse. Example: `https://memo.vaaat.com`

---

## Feedback & Community

- Report bugs via [Issues](https://github.com/lin-snow/Ech0/issues).
- Propose features or share ideas in [Discussions](https://github.com/lin-snow/Ech0/discussions).

---

## Architecture

![Architecture Diagram](./docs/imgs/Ech0技术架构图.svg)  
> by ExcaliDraw

---

## Development Guide

### Backend Requirements
- Go 1.25.3+  
- C Compiler for CGO (`go-sqlite3`):
  - Windows: [MinGW-w64](https://winlibs.com/)  
  - macOS: `brew install gcc`  
  - Linux: `sudo apt install build-essential`  
- Google Wire: `go install github.com/google/wire/cmd/wire@latest`  
- Golangci-Lint: `golangci-lint run` / `golangci-lint fmt`  
- Swagger: `swag init -g internal/server/server.go -o internal/swagger`  

### Frontend Requirements
- NodeJS v24.10.0+, PNPM v10.20.1+  
- Use [fnm](https://github.com/Schniz/fnm) if multiple Node versions needed

### Start Backend & Frontend
```shell
# Backend
go run main.go

# Frontend
cd web
pnpm install
pnpm dev
```

Preview: Backend `http://localhost:6277`, Frontend `http://localhost:5173`

> When importing layered packages, prefer consistent aliases such as `xxxModel`, `xxxService`, `xxxRepository`, and so on.


---

## Thanks for Your Support!

Thank you to all the friends who have supported this project! Your contributions keep it thriving 💡✨

|                        ⚙️ User                        |   🔋 Date   | 💬 Message                                       |
| :--------------------------------------------------: | :--------: | :---------------------------------------------- |
|                  🧑‍💻 Anonymous Friend                  | 2025-5-19  | Silly programmer, buy yourself some sweet drink |
|        🧑‍💻 [@sseaan](https://github.com/sseaan)        | 2025-7-27  | Ech0 is a great thing🥳                          |
| 🧑‍💻 [@QYG2297248353](https://github.com/QYG2297248353) | 2025-10-10 | None                                            |
|    🧑‍💻 [@continue33](https://github.com/continue33)    | 2025-10-23 | Thanks for fixing R2                            |
|    🧑‍💻 [@hoochanlon](https://github.com/hoochanlon)      | 2025-10-28 | None             |
|       🧑‍💻 [@Rvn0xsy](https://github.com/Rvn0xsy)       | 2025-11-12 | Great project, I will keep following! |
|                     🧑‍💻 王贼臣                     | 2025-11-20 | Thanks www.cardopt.cn             |
|       🧑‍💻 [@ljxme](https://github.com/ljxme)    | 2025-11-30 | Doing my humble part 😋             |
|       🧑‍💻 [@he9ab2l](https://github.com/he9ab2l)    | 2025-12-23 | None            |
|       🧑‍💻 鸿运当头(windfore)    | 2026-1-6 | Thank you for creating ech0           |

---

## Star History

<a href="https://www.star-history.com/#lin-snow/Ech0&Timeline">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=lin-snow/Ech0&type=Timeline&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=lin-snow/Ech0&type=Timeline" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=lin-snow/Ech0&type=Timeline" />
 </picture>
</a>


---

## Acknowledgements

- Thanks to all users for their valuable suggestions and feedback.
- Thanks to all contributors and supporters from the open-source community.


![Alt](https://repobeats.axiom.co/api/embed/d69b9177e4a121e31aaed95354ff862c928ca22d.svg "Repobeats analytics image")

---

## Support

🌟 If you like **Ech0**, please give it a Star! 🚀  
Ech0 is completely free and open-source. Support helps the project continue improving.  

|                  Platform                  | QR Code                                                |
| :----------------------------------------: | :----------------------------------------------------- |
| [**Afdian**](https://afdian.com/a/l1nsn0w) | <img src="./docs/imgs/pay.jpeg" alt="Pay" width="200"> |

---

```cpp

███████╗     ██████╗    ██╗  ██╗     ██████╗ 
██╔════╝    ██╔════╝    ██║  ██║    ██╔═████╗
█████╗      ██║         ███████║    ██║██╔██║
██╔══╝      ██║         ██╔══██║    ████╔╝██║
███████╗    ╚██████╗    ██║  ██║    ╚██████╔╝
╚══════╝     ╚═════╝    ╚═╝  ╚═╝     ╚═════╝ 

``` 
