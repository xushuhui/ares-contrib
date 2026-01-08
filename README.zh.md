# Ares Contrib

[English](README.md)

[Ares](https://github.com/xushuhui/ares) Web 框架的扩展中间件集合 - 一个基于 chi 路由构建的轻量级、高性能 Go Web 框架。

## 📋 目录

- [项目概述](#项目概述)
- [安装](#安装)
- [可用中间件](#可用中间件)
- [快速开始](#快速开始)
- [中间件示例](#中间件示例)
- [最佳实践](#最佳实践)
- [测试](#测试)
- [基准测试](#基准测试)
- [依赖项](#依赖项)
- [贡献指南](#贡献指南)
- [许可证](#许可证)

## 🎯 项目概述

Ares Contrib 提供了一套生产就绪的中间件集合，扩展了 Ares 框架的功能，同时保持核心框架的轻量级特性。每个中间件都具备：

- ✅ **充分测试** - 全面的测试覆盖（总体 87%+）
- ✅ **生产就绪** - 已在实际项目中使用
- ✅ **高性能** - 针对速度和内存效率进行了优化
- ✅ **灵活配置** - 使用函数式选项模式
- ✅ **标准库兼容** - 遵循 Go 最佳实践

## 📦 安装

```bash
go get github.com/xushuhui/ares-contrib
```

## 🚀 可用中间件

### 中间件概览

| 中间件 | 覆盖率 | 描述 | 状态 |
|--------|--------|------|------|
| [RequestID](#request-id) | 100% | 唯一请求追踪 | ✅ 稳定 |
| [Secure](#安全头) | 100% | 安全头保护 | ✅ 稳定 |
| [CORS](#cors) | 96.2% | 跨域资源共享 | ✅ 稳定 |
| [JWT](#jwt-认证) | 85.7% | 令牌认证 | ✅ 稳定 |
| [GZIP](#gzip-压缩) | 80.9% | 响应压缩 | ✅ 稳定 |
| [BodyLimit](#请求体限制) | 72.7% | 请求体大小限制 | ✅ 稳定 |
| [RateLimiter](#限流器) | 72.0% | 基于 IP/密钥的限流 | ✅ 稳定 |

---

## 🔥 快速开始

```go
package main

import (
    "github.com/xushuhui/ares"
    "github.com/xushuhui/ares-contrib/middleware/cors"
    "github.com/xushuhui/ares-contrib/middleware/gzip"
    "github.com/xushuhui/ares-contrib/middleware/requestid"
    "github.com/xushuhui/ares-contrib/middleware/secure"
    "github.com/xushuhui/ares-contrib/middleware/jwt"
)

func main() {
    app := ares.New()

    // 添加中间件
    app.Use(requestid.New())
    app.Use(secure.New())
    app.Use(cors.New(
        cors.WithAllowedOrigins([]string{"https://example.com"}),
        cors.WithAllowCredentials(true),
    ))
    app.Use(gzip.New(gzip.WithLevel(5)))

    // 公开路由
    app.POST("/login", loginHandler)

    // JWT 保护的受控路由
    api := app.Group("/api", jwt.New([]byte("your-secret-key")))
    api.GET("/users", getUsersHandler)
    api.GET("/profile", getProfileHandler)

    app.Run(":8080")
}
```

---

## 📚 中间件示例

### Request ID

生成唯一的请求 ID，用于分布式追踪和日志记录。

**特性：**
- 默认使用 UUID v4 生成
- 支持自定义生成器
- 重用请求头中已有的 ID
- 基于上下文访问

**使用方法：**

```go
import "github.com/xushuhui/ares-contrib/middleware/requestid"

// 默认配置
app.Use(requestid.New())

// 自定义配置
app.Use(requestid.New(
    requestid.WithGenerator(func() string {
        return "req-" + uuid.New().String()
    }),
    requestid.WithHeader("X-Request-ID"),
    requestid.WithContextKey("request_id"),
))

// 在处理器中访问
app.GET("/test", func(ctx *ares.Context) error {
    reqID := ctx.GetString("request_id")
    ctx.Logger().Info("处理请求", "id", reqID)
    return ctx.JSON(200, map[string]string{"request_id": reqID})
})
```

**响应头：**
```
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
```

---

### 安全头

使用安全头保护免受常见 Web 漏洞攻击。

**特性：**
- XSS 防护
- Content Type Options (nosniff)
- X-Frame-Options（点击劫持防护）
- HSTS（HTTP 严格传输安全）
- 内容安全策略
- Referrer 策略
- 权限策略

**使用方法：**

```go
import "github.com/xushuhui/ares-contrib/middleware/secure"

// 默认配置
app.Use(secure.New())

// 生产环境配置
app.Use(secure.New(
    secure.WithXSSProtection("1; mode=block"),
    secure.WithContentTypeNosniff("nosniff"),
    secure.WithXFrameOptions("DENY"),
    secure.WithHSTSMaxAge(31536000),           // 1 年
    secure.WithHSTSIncludeSubdomains(true),
    secure.WithContentSecurityPolicy(
        "default-src 'self'; " +
        "script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
        "style-src 'self' 'unsafe-inline'; " +
        "img-src 'self' data: https:; " +
        "font-src 'self' data:;",
    ),
    secure.WithReferrerPolicy("strict-origin-when-cross-origin"),
    secure.WithPermissionsPolicy(
        "geolocation=(self), " +
        "microphone=(), " +
        "camera=(), " +
        "payment=()",
    ),
))
```

**默认头：**
```
X-XSS-Protection: 1; mode=block
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
```

**最佳实践：**
- 仅在启用 HTTPS 时启用 HSTS
- 先使用 CSP report-only 模式测试策略
- 根据需要定期更新 CSP 策略

---

### CORS

跨域资源共享中间件，用于 API 访问控制。

**特性：**
- 可配置的允许来源、方法、头
- 支持凭证
- 预检请求处理（OPTIONS）
- Max-age 配置
- 自动管理 Vary 头

**使用方法：**

```go
import "github.com/xushuhui/ares-contrib/middleware/cors"

// 允许所有来源（仅用于开发！）
app.Use(cors.New())

// 生产环境配置
app.Use(cors.New(
    cors.WithAllowedOrigins([]string{
        "https://example.com",
        "https://www.example.com",
    }),
    cors.WithAllowedMethods([]string{
        "GET", "POST", "PUT", "DELETE", "OPTIONS",
    }),
    cors.WithAllowedHeaders([]string{
        "Authorization",
        "Content-Type",
        "X-Requested-With",
    }),
    cors.WithExposedHeaders([]string{
        "X-Total-Count",
        "X-Page-Count",
    }),
    cors.WithAllowCredentials(true),
    cors.WithMaxAge(3600), // 1 小时
))

// API 端点
app.GET("/api/data", handler)
```

**CORS 与凭证：**
```go
// ❌ 错误：不能在凭证模式下使用通配符
app.Use(cors.New(
    cors.WithAllowCredentials(true),
))

// ✅ 正确：使用凭证时指定来源
app.Use(cors.New(
    cors.WithAllowedOrigins([]string{"https://example.com"}),
    cors.WithAllowCredentials(true),
))
```

**最佳实践：**
- 永远不要在 `AllowCredentials: true` 时使用通配符（`*`）
- 在生产环境中明确指定允许的来源
- 设置合适的 MaxAge 以减少预检请求

---

### JWT 认证

基于 JWT 的令牌认证中间件。

**特性：**
- 支持多种签名算法（HS256、HS512 等）
- 自定义声明支持
- 详细的错误分类
- JSON 错误响应
- 基于上下文的声明存储

**使用方法：**

```go
import (
    "github.com/golang-jwt/jwt/v5"
    "github.com/xushuhui/ares-contrib/middleware/jwt"
)

// 简单使用
app.Use(jwt.New([]byte("your-secret-key")))

// 使用自定义声明
type CustomClaims struct {
    UserID   string `json:"user_id"`
    Email    string `json:"email"`
    jwt.RegisteredClaims
}

api := app.Group("/api", jwt.New(
    []byte("your-secret-key"),
    jwt.WithSigningMethod(jwt.SigningMethodHS256),
    jwt.WithClaims(func() jwt.Claims {
        return &CustomClaims{}
    }),
    jwt.WithContextKey("user"),
))

// 在处理器中访问声明
api.GET("/profile", func(ctx *ares.Context) error {
    claims, ok := jwt.GetClaims(ctx.Request.Context())
    if !ok {
        return ctx.JSON(401, map[string]string{"error": "未授权"})
    }

    customClaims, ok := claims.(*CustomClaims)
    if !ok {
        return ctx.JSON(500, map[string]string{"error": "无效的声明类型"})
    }

    return ctx.JSON(200, map[string]interface{}{
        "user_id": customClaims.UserID,
        "email":   customClaims.Email,
    })
})
```

**创建令牌：**

```go
func generateToken(userID string) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Hour * 24).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte("your-secret-key"))
}
```

**错误响应：**

```json
// 缺少令牌
{
  "code": 401,
  "message": "JWT token is missing"
}

// 过期的令牌
{
  "code": 401,
  "message": "JWT token has expired"
}

// 无效的令牌
{
  "code": 401,
  "message": "token is invalid"
}
```

**最佳实践：**
- 在环境变量中存储密钥
- 使用强随机密钥（256+ 位）
- 实现令牌刷新机制
- 每次请求时验证令牌过期时间

---

### GZIP 压缩

响应压缩中间件，减少带宽使用。

**特性：**
- 可配置的压缩级别（1-9）
- 最小响应大小阈值
- 排除特定文件扩展名
- 排除特定路径（WebSocket、流）
- Writer 池化以提高性能

**使用方法：**

```go
import "github.com/xushuhui/ares-contrib/middleware/gzip"

// 默认配置
app.Use(gzip.New())

// 自定义配置
app.Use(gzip.New(
    gzip.WithLevel(5),                        // 压缩级别（1-9）
    gzip.WithMinLength(1024),                 // 仅压缩 > 1KB 的响应
    gzip.WithExcludedExtensions([]string{
        ".png", ".jpg", ".jpeg", ".gif",     // 已压缩的文件
        ".zip", ".gz", ".tar",
        ".pdf", ".mp4", ".mp3",
    }),
    gzip.WithExcludedPaths([]string{
        "/api/stream",                        // WebSocket/流
        "/ws",
        "/download",
    }),
))
```

**默认排除的扩展名：**
- 图片：`.png`、`.jpg`、`.jpeg`、`.gif`、`.webp`、`.svg`
- 压缩包：`.zip`、`.gz`、`.tar`、`.rar`、`.7z`
- 媒体：`.mp4`、`.avi`、`.mov`、`.mp3`、`.wav`
- 文档：`.pdf`

**性能提示：**
- 级别 5-7 提供了良好的平衡
- 不要压缩已压缩的文件（图片、视频）
- 排除流式端点
- 监控 CPU 使用率与带宽节省的对比

---

### 请求体限制

限制请求体大小，防止内存耗尽攻击。

**特性：**
- 可配置的大小限制
- 使用 `http.MaxBytesReader` 高效限制
- 超出时返回 413 Payload Too Large

**使用方法：**

```go
import "github.com/xushuhui/ares-contrib/middleware/bodylimit"

// 全局限制：10MB
app.Use(bodylimit.New(10 * 1024 * 1024))

// 不同路由使用不同限制
uploadGroup := app.Group("/upload", bodylimit.New(100 * 1024 * 1024)) // 100MB
uploadGroup.POST("/image", uploadImageHandler)

apiGroup := app.Group("/api", bodylimit.New(1 * 1024 * 1024)) // 1MB
apiGroup.POST("/data", postDataHandler)
```

**错误响应：**
```
HTTP 413 Payload Too Large
```

**最佳实践：**
- 为 API 端点设置较低的限制
- 为文件上传设置较高的限制
- 考虑对大文件使用分块上传

---

### 限流器

令牌桶限流器，防止 API 滥用。

**特性：**
- 默认基于 IP 限流
- 自定义密钥提取（用户 ID、API 密钥）
- 可配置的速率和突发
- 自动清理旧的限流器
- 自定义错误处理器支持

**使用方法：**

```go
import "github.com/xushuhui/ares-contrib/middleware/ratelimiter"

// 默认：每秒 10 个请求，突发 20 个
app.Use(ratelimiter.New())

// 自定义配置
app.Use(ratelimiter.New(
    ratelimiter.WithRate(100),              // 每秒 100 个请求
    ratelimiter.WithBurst(200),             // 允许突发 200 个
    ratelimiter.WithKeyFunc(func(r *http.Request) string {
        // 按用户 ID 限流而不是 IP
        userID := r.Header.Get("X-User-ID")
        if userID != "" {
            return "user:" + userID
        }
        return "ip:" + r.RemoteAddr
    }),
    ratelimiter.WithErrorHandler(func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, "超过速率限制。请稍后再试。", 429)
    }),
))
```

**为不同路由设置不同限制：**

```go
// 公共 API：每秒 10 个请求
publicAPI := app.Group("/api/public", ratelimiter.New(
    ratelimiter.WithRate(10),
))

// 认证用户：每秒 100 个请求
userAPI := app.Group("/api/user", ratelimiter.New(
    ratelimiter.WithRate(100),
    ratelimiter.WithKeyFunc(func(r *http.Request) string {
        return r.Context().Value("user_id").(string)
    }),
))
```

**最佳实践：**
- 为公共用户和认证用户设置不同的限制
- 考虑突发容量以提升用户体验
- 根据使用模式监控和调整
- 实现指数退避的请求重试

---

## 🎯 最佳实践

### 中间件顺序

中间件的顺序很重要！推荐的顺序如下：

```go
app := ares.New()

// 1. Request ID（第一个，用于追踪）
app.Use(requestid.New())

// 2. 安全头（尽早设置）
app.Use(secure.New())

// 3. 限流（在昂贵的操作之前）
app.Use(ratelimiter.New())

// 4. 请求体限制（在读取请求体之前）
app.Use(bodylimit.New(10 * 1024 * 1024))

// 5. CORS（在认证之前）
app.Use(cors.New())

// 6. 压缩（在响应之前）
app.Use(gzip.New())

// 7. 认证（用于受保护的路由）
api := app.Group("/api", jwt.New(secret))
```

### 性能优化提示

1. **仅对基于文本的内容使用 GZIP**
   ```go
   gzip.WithExcludedExtensions([]string{".png", ".jpg", ".mp4"})
   ```

2. **设置合适的速率限制**
   ```go
   // 太低：用户体验差
   ratelimiter.WithRate(1)  // ❌

   // 太高：没有保护
   ratelimiter.WithRate(10000)  // ❌

   // 适度
   ratelimiter.WithRate(100)  // ✅
   ```

3. **使用上下文而不是全局变量**
   ```go
   // ❌ 错误
   var userID string

   // ✅ 正确
   userID := ctx.GetString("user_id")
   ```

### 安全清单

- [ ] 在生产环境中启用 HTTPS
- [ ] 设置安全 Cookie
- [ ] 实施限流保护
- [ ] 使用 CSP 防止 XSS
- [ ] 启用较长时间的 HSTS
- [ ] 验证和清理输入
- [ ] 保持依赖项更新
- [ ] 记录安全事件
- [ ] 实施身份认证
- [ ] 正确使用 CORS（不要在凭证模式下使用通配符）

---

## 🧪 测试

所有中间件都有全面的测试覆盖：

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test ./... -cover

# 生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**测试覆盖率总结：**

```
中间件              覆盖率      测试数量
----------------------------------------
RequestID           100.0%      6
Secure              100.0%      11
CORS                96.2%       14
JWT                 85.7%       10
GZIP                80.9%       14
BodyLimit           72.7%       8
RateLimiter         72.0%       6
----------------------------------------
总计                ~87%        69
```

---

## 📊 基准测试

运行基准测试以测试中间件性能：

```bash
cd middleware/<middleware-name>
go test -bench=. -benchmem
```

示例结果（Apple M1 Pro, Go 1.23）：

```
BenchmarkRequestID-8       10000000    105 ns/op    0 B/op    0 allocs/op
BenchmarkSecure-8          10000000    120 ns/op    0 B/op    0 allocs/op
BenchmarkCORS-8            5000000     250 ns/op    0 B/op    0 allocs/op
BenchmarkJWT-8            1000000     1200 ns/op   512 B/op  8 allocs/op
BenchmarkGZIP-8           3000000     450 ns/op    128 B/op  2 allocs/op
```

---

## 📦 依赖项

| 包 | 版本 | 用途 |
|---------|---------|---------|
| [github.com/golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) | ^5.2.0 | JWT 实现 |
| [github.com/google/uuid](https://github.com/google/uuid) | ^1.5.0 | UUID 生成 |
| [golang.org/x/time/rate](https://golang.org/x/time/rate) | latest | 限流 |
| [github.com/xushuhui/ares](https://github.com/xushuhui/ares) | latest | 核心框架 |

---

## 🤝 贡献指南

我们欢迎贡献！请遵循以下步骤：

1. Fork 本仓库
2. 创建功能分支（`git checkout -b feature/amazing-feature`）
3. 为您的更改编写测试
4. 确保所有测试通过（`go test ./...`）
5. 保持测试覆盖率在 80% 以上
6. 提交您的更改（`git commit -m 'Add amazing feature'`）
7. 推送到分支（`git push origin feature/amazing-feature`）
8. 创建 Pull Request

**开发要求：**
- Go 1.21 或更高版本
- 遵循 Go 最佳实践和 Effective Go 指南
- 编写清晰、符合 Go 惯用法代码
- 为新功能包含测试
- 根据需要更新文档

---

## 📄 许可证

本项目在 MIT 许可证下授权 - 详见 [LICENSE](LICENSE) 文件。

---

## 🔗 相关链接

- [Ares 框架](https://github.com/xushuhui/ares) - 核心框架
- [文档](https://github.com/xushuhui/ares/wiki) - 官方文档
- [示例](./examples/) - 使用示例
- [问题反馈](https://github.com/xushuhui/ares-contrib/issues) - Bug 报告和功能请求

---

## 🌟 Star 历史

如果你觉得这个项目有用，请考虑给它一个 ⭐ star！

---

由 Ares 社区用 ❤️ 制作
