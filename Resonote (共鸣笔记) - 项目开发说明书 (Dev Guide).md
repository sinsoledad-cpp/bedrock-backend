# Resonote (共鸣笔记) - 项目开发说明书

## 0. 文档概览

- **项目名称**：Resonote（共鸣笔记）
- **文档类型**：项目开发说明书 (Development Guide)
- **版本号**：v1.0.0
- **适用阶段**：V1 MVP 开发与架构设计
- **技术栈核心**：Golang (单体) + Vue 3 + Ant Design Vue

### 0.1 产品背景与目标
- **核心理念**：结合 AI 情感计算与即时通讯，打造基于「情绪共鸣」的私密社交空间。
- **核心价值**：
  - **智能日记**：AI 辅助表达与整理情绪。
  - **灵魂共鸣**：基于情绪向量匹配陌生人。
  - **AI 社交**：破冰与润滑社交关系。
- **开发目标**：构建高性能、易维护的单体应用，快速验证产品 MVP。

---

## 1. 整体技术架构

采用 **前后端分离** 架构，后端采用 **Golang 单体** 模式以保证性能与开发效率，前端采用 **Vue 3** 生态。

### 1.1 技术选型

#### 前端 (Web Client)
- **核心框架**：Vue 3 (Composition API) + Vite
- **UI 组件库**：Ant Design Vue (v4.x)
- **状态管理**：Pinia
- **路由管理**：Vue Router
- **网络请求**：Axios
- **实时通信**：Native WebSocket
- **CSS 方案**：Unocss 或 TailwindCSS (可选，配合 AntDV)

#### 后端 (Server)
- **开发语言**：Golang (Go 1.21+)
- **Web 框架**：Gin (轻量、高性能) 或 Echo
- **ORM 框架**：GORM (v2) 或 Ent (类型安全)
- **配置管理**：Viper
- **日志库**：Zap + Lumberjack
- **API 文档**：Swagger (Swag)

#### 基础设施 & 中间件
- **数据库**：MySQL 8.0 或 PostgreSQL 15+ (存储业务数据)
- **缓存/消息**：Redis 7.x (Session、验证码、限流、简单的消息队列)
- **向量数据库**：Pgvector (若用 PG) 或 Milvus/Qdrant (独立部署，用于情绪向量检索)
- **对象存储**：MinIO 或 AWS S3 (存储图片/头像)

### 1.2 架构图示 (逻辑分层)

```mermaid
graph TD
    User[用户 (Web/H5)] --> Nginx[Nginx 网关]
    Nginx --> Frontend[前端静态资源 (Vue3 + AntDV)]
    Nginx --> Backend[Golang 单体服务 (API + WS)]
    
    subgraph "Golang Monolith"
        API[API Layer (Gin)]
        WS[WebSocket Manager]
        Service[Service Layer (Biz Logic)]
        Model[Data Access Layer (GORM)]
    end
    
    Backend --> DB[(MySQL/PostgreSQL)]
    Backend --> Redis[(Redis Cache)]
    Backend --> VectorDB[(Vector DB)]
    Backend --> AI[外部 AI 服务 (LLM / Image Gen)]
```

---

## 2. 后端工程结构规范

采用 Golang 标准项目结构 (Standard Go Project Layout) 的变体，适合单体应用。

```text
resonote-server/
├── cmd/
│   └── server/
│       └── main.go           # 程序入口
├── configs/                  # 配置文件 (config.yaml)
├── internal/                 # 私有代码，不对外暴露
│   ├── api/                  # HTTP Handler (Controller)
│   │   ├── v1/
│   │   │   ├── auth.go
│   │   │   ├── journal.go
│   │   │   └── resonance.go
│   │   └── router.go         # 路由定义
│   ├── core/                 # 核心组件 (Logger, DB, Redis 初始化)
│   ├── model/                # 数据库模型 (Struct)
│   ├── repository/           # 数据访问层 (DAO)
│   ├── service/              # 业务逻辑层
│   ├── middleware/           # 中间件 (JWT, CORS, RateLimit)
│   ├── pkg/                  # 内部工具包 (Utils, AI Client)
│   └── websocket/            # WebSocket 核心逻辑 (Hub, Client)
├── pkg/                      # 可导出的公共库 (如有)
├── docs/                     # Swagger 文档
├── go.mod
└── go.sum
```

---

## 3. 核心模块详细设计

### 3.1 用户与权限模块 (User & Auth)

- **功能**：注册、登录 (邮箱/手机/OAuth)、个人主页、隐私设置。
- **技术实现**：
  - 密码存储：使用 `bcrypt` 进行哈希加密。
  - 鉴权：使用 `JWT (JSON Web Token)`，包含 `Access Token` (短效) 和 `Refresh Token` (长效)。
  - 登录保护：Redis 计数器实现错误次数限制，触发图形验证码。

### 3.2 智能日记模块 (Smart Journal)

- **功能**：CRUD 日记、Markdown 解析、AI 辅助 (扩写/润色/摘要/配图)。
- **技术实现**：
  - **存储**：日记正文存 `TEXT/LONGTEXT`。
  - **AI 流程**：
    1. 用户提交内容 -> 存入 DB (Status: Processing)。
    2. Go 协程异步调用 LLM API 生成摘要、标签、向量。
    3. Go 协程异步调用绘图 API 生成封面。
    4. 回调或轮询更新 DB (Status: Published)。
  - **向量化**：调用 Embedding API 将文本转为向量，存入向量库。

### 3.3 共鸣社交模块 (Resonance)

- **功能**：灵魂匹配、匿名共鸣墙、共鸣点赞。
- **技术实现**：
  - **匹配算法**：基于当前用户最新日记的向量，在向量库中进行 `Cosine Similarity` (余弦相似度) 搜索。
  - **推荐策略**：`Score = 相似度 * 0.7 + 时间衰减因子 * 0.3`，优先推荐近期的高共鸣内容。
  - **共鸣墙**：Redis ZSet 维护热度榜单 (可选)，MySQL 存储点赞记录防重。

### 3.4 实时通信模块 (IM & Rooms)

- **功能**：临时房间、单聊、状态同步、AI 破冰。
- **技术实现**：
  - **协议**：WebSocket。
  - **并发模型**：每个 WebSocket 连接对应两个 Goroutine (ReadPump, WritePump)。
  - **房间管理**：内存中维护 `Hub` 结构，包含 `map[roomId]map[*Client]bool`。
  - **房间销毁**：使用 `time.AfterFunc` 或 `Redis Key Expiration` (Key过期通知) 触发房间清理逻辑。
  - **AI 破冰**：每个房间维护 `LastMessageTime`，后台 Goroutine 轮询 (Ticker) 检查超时房间，触发 AI 发言。

---

## 4. 数据库设计概览

### 4.1 用户表 (users)
| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| id | BIGINT | 主键, Snowflake ID |
| email | VARCHAR | 邮箱 (唯一索引) |
| password_hash | VARCHAR | 密码 |
| nickname | VARCHAR | 昵称 |
| role | TINYINT | 0:Visitor, 1:Member, 2:Premium, 9:Admin |
| ai_persona | TEXT | AI 人格描述 |
| settings | JSON | 隐私设置、偏好 |

### 4.2 日记表 (journals)
| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| id | BIGINT | 主键 |
| user_id | BIGINT | 关联用户 |
| content | TEXT | 日记内容 |
| visibility | TINYINT | 0:Private, 1:Friends, 2:Anonymous |
| emotion_vector | VECTOR(768) | 情绪向量 (需数据库支持或独立存) |
| emotion_tags | JSON | 标签数组 ["焦虑", "深夜"] |
| cover_url | VARCHAR | AI 封面图链接 |

### 4.3 房间表 (rooms)
| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| id | BIGINT | 主键 |
| invite_code | CHAR(6) | 邀请码 (索引) |
| topic | VARCHAR | 主题 |
| expire_at | TIMESTAMP | 过期时间 |
| status | TINYINT | 1:Active, 0:Expired |

### 4.4 消息表 (messages)
| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| id | BIGINT | 主键 |
| room_id | BIGINT | 关联房间/会话 |
| sender_id | BIGINT | 发送者 (0代表系统/AI) |
| type | TINYINT | 1:Text, 2:Image, 3:Card |
| content | TEXT | 内容或JSON载荷 |

---

## 5. 接口规范

遵循 RESTful 风格，响应格式统一：

```json
{
  "code": 200,      // 业务状态码 (200:成功, >200:错误)
  "msg": "success", // 提示信息
  "data": { ... }   // 业务数据
}
```

### 核心接口示例

- **Auth**: `POST /api/v1/auth/login`, `POST /api/v1/auth/register`
- **Journal**: 
  - `POST /api/v1/journals` (发布)
  - `GET /api/v1/journals` (列表)
  - `POST /api/v1/journals/ai/expand` (AI扩写)
- **Resonance**: 
  - `GET /api/v1/resonance/match` (灵魂匹配)
  - `GET /api/v1/resonance/wall` (匿名墙)
- **Room**: 
  - `POST /api/v1/rooms` (创建)
  - `POST /api/v1/rooms/join` (加入)

---

## 6. 非功能性开发要求

1.  **性能优化**：
    -   API 响应时间目标 < 100ms (非 AI 接口)。
    -   WebSocket 单机并发目标 > 1000 连接。
2.  **安全性**：
    -   敏感数据 (日记) 考虑应用层加密存储 (AES-GCM)。
    -   API 限流 (Rate Limiting)：使用 `golang.org/x/time/rate` 或 Redis 中间件。
3.  **代码规范**：
    -   遵循 `Uber Go Style Guide`。
    -   前端遵循 Vue3 官方风格指南 (Priority A & B)。
4.  **错误处理**：
    -   后端统一 Error Handler，避免 Panic 导致服务崩溃。
    -   前端统一 Axios Interceptor 处理 401/403/500 错误。

---

## 7. 部署与运维

-   **构建**：
    -   后端：`go build -o server cmd/server/main.go` (多阶段 Docker 构建)。
    -   前端：`npm run build` -> 生成 `dist` 目录。
-   **部署模式**：
    -   Docker Compose (推荐开发/测试环境)：编排 App, MySQL, Redis。
    -   生产环境：Nginx 反向代理 -> Go 服务 (端口 8080)。
-   **CI/CD**：
    -   GitLab CI / GitHub Actions 自动运行 `go test` 和 `npm run lint`。
