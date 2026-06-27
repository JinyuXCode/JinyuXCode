<!--
  ┌─────────────────────────────────────────────────────────────┐
  │  service: jinyux                                              │
  │  concept: a frontend dev, deployed as a full-stack service   │
  └─────────────────────────────────────────────────────────────┘
-->

<div align="center">

<img src="https://readme-typing-svg.demolab.com?font=JetBrains+Mono&weight=700&size=26&pause=900&color=00ADD8&center=true&vCenter=true&width=620&lines=%24+systemctl+status+jinyux;%E2%97%8F+active+(running)+%E2%80%94+shipping+daily;frontend+%E2%86%92+full-stack%2C+deploying+myself" alt="boot" />

<br/><br/>

<!-- live status line -->
<img src="https://img.shields.io/badge/status-%E2%97%8F_active_(running)-00ADD8?style=for-the-badge&labelColor=0d1117" />
&nbsp;
<img src="https://img.shields.io/badge/role-frontend_engineer-1f2937?style=for-the-badge&labelColor=0d1117" />
&nbsp;
<img src="https://img.shields.io/badge/region-cn--zhengzhou-1f2937?style=for-the-badge&labelColor=0d1117" />
&nbsp;
<img src="https://img.shields.io/badge/uptime-keep_shipping-1f2937?style=for-the-badge&labelColor=0d1117" />

</div>

---

### `▸ GET /whoami`

> 一名前端工程师，把自己当成一个正在迭代上线的服务。
> A frontend engineer treating myself as a service in continuous deployment.

```go
// jinyux exposes a single, honest endpoint.
func (dev *Developer) Whoami(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(Profile{
        Name:     "JinyuX",
        Role:     "Frontend Engineer",          // 主业，做得扎实
        Location: "Zhengzhou, China 🇨🇳",
        Shipping: []string{"Vue 3", "TypeScript", "React"},
        Learning: "Go",                          // 正在补齐后端
        Pairing:  []string{"Cursor", "Claude Code"}, // AI 辅助工作流
        Goal:     "Full-Stack — owning a feature end to end 🚀",
    })
    // 200 OK · always building
}
```

---

### `▸ ARCHITECTURE` · 技术栈 / The Stack

我按真实的请求路径组织技术栈 —— 从浏览器到服务端，目前主力在上层，正向下层延伸。
*Organized the way a request actually travels: strongest at the top, growing downward.*

```text
   ┌─ CLIENT ─────────────────────────────────────┐
   │  Vue 3 · TypeScript · React · Vite · UnoCSS   │   ███████████  主力 / primary
   └───────────────────────────────────────────────┘
                        │  HTTP / JSON
                        ▼
   ┌─ EDGE & TOOLING ─────────────────────────────┐
   │  Node.js · Git · Cursor · Claude Code         │   ███████░░░░  日常 / daily
   └───────────────────────────────────────────────┘
                        │  learning…
                        ▼
   ┌─ SERVER ─────────────────────────────────────┐
   │  Go (net/http → services → deploy)            │   ████░░░░░░░  扩展中 / expanding
   └───────────────────────────────────────────────┘
```

<div align="center">

![Vue](https://img.shields.io/badge/Vue_3-35495E?style=flat-square&logo=vuedotjs&logoColor=4FC08D)
![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=flat-square&logo=typescript&logoColor=white)
![React](https://img.shields.io/badge/React-20232A?style=flat-square&logo=react&logoColor=61DAFB)
![Vite](https://img.shields.io/badge/Vite-646CFF?style=flat-square&logo=vite&logoColor=FFD62E)
![UnoCSS](https://img.shields.io/badge/UnoCSS-333333?style=flat-square&logo=unocss&logoColor=white)
&nbsp;·&nbsp;
![Go](https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white)
![Node.js](https://img.shields.io/badge/Node.js-339933?style=flat-square&logo=nodedotjs&logoColor=white)
&nbsp;·&nbsp;
![Git](https://img.shields.io/badge/Git-F05032?style=flat-square&logo=git&logoColor=white)
![Cursor](https://img.shields.io/badge/Cursor-000000?style=flat-square&logo=cursor&logoColor=white)

</div>

---

### `▸ DEPLOY PIPELINE` · 学习路线 / Roadmap

把「转全栈」当成一条流水线来跑 —— 每个 stage 都有明确的产出。
*Treating "go full-stack" as a pipeline — every stage ships something real.*

```text
 stage          progress                    state
 ─────────────────────────────────────────────────────────
 frontend       ████████████████████  100%  ✓ passed
 typescript     ██████████████████░░   90%  ✓ passed
 react          ██████████████░░░░░░   70%  ⟳ running
 go             ██████░░░░░░░░░░░░░░   30%  ⟳ running
 full-stack     ████░░░░░░░░░░░░░░░░   20%  ⧗ queued  ← target
```

📦 **Artifacts** — 路线与进度都已落盘，可点开追踪：
- 🗺️ [**Go 全栈进阶路线图**](./Go-Fullstack-Roadmap.md) — 六阶段，从 CLI 到上线
- ✅ [**学习任务清单**](./Go-Learning-Tasks.md) — 可勾选的逐项任务，实时追踪进度

---

### `▸ TELEMETRY` · 运行数据 / Stats

<div align="center">

<img height="165" src="https://github-readme-stats.vercel.app/api?username=JinyuXCode&show_icons=true&hide_border=true&title_color=00ADD8&icon_color=00ADD8&text_color=8b949e&bg_color=0d1117" />
<img height="165" src="https://github-readme-stats.vercel.app/api/top-langs/?username=JinyuXCode&layout=compact&hide_border=true&title_color=00ADD8&text_color=8b949e&bg_color=0d1117" />

<br/>

<img src="https://github-readme-streak-stats.herokuapp.com/?user=JinyuXCode&theme=transparent&hide_border=true&stroke=00ADD8&ring=00ADD8&fire=00ADD8&currStreakLabel=00ADD8" width="86%" />

</div>

---

<div align="center">

```text
$ tail -f /var/log/jinyux.log
[INFO] frontend stable · backend warming up · pipeline green
[INFO] next: deploy myself, full-stack 🚀
```

<sub>「保持构建，持续交付。」 · <b>Keep shipping.</b></sub>

</div>
