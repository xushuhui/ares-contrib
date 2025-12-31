# Ares Contrib

Ares Web 框架的扩展中间件和工具集。

此包包含扩展 Ares 功能的额外中间件，但与核心框架分离以保持其轻量级特性。

## 安装

```bash
go get github.com/xushuhui/ares-contrib
```

## 可用中间件

### CORS

跨域资源共享中间件。

**特性：**
- 可配置的允许来源、方法和头
- 支持凭证
- 预检请求处理
- 最大年龄配置

**使用方法：**

```go
import (
    "github.com/xushuhui/ares"
    "github.com/xushuhui/ares-contrib/middleware"
)

func main() {
    app := ares.New()

    // 使用默认配置的简单用法
    app.Use(middleware.CORS(middleware.DefaultCORSOptions))

    // 自定义配置
    app.Use(middleware.CORS(middleware.CORSOptions{
        AllowedOrigins:   []string{"https://example.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Authorization", "Content-Type"},
        AllowCredentials: true,
        MaxAge:           3600,
    }))

    app.Run(":8080")
}
```

### Request ID

生成唯一的请求 ID 用于跟踪和日志记录。

**特性：**
- 默认使用 UUID v4 生成
- 支持自定义 ID 生成器
- 重用头中现有的请求 ID
- 将 ID 存储在上下文中以便在处理器中访问

**使用方法：**

```go
import (
    "github.com/xushuhui/ares"
    "github.com/xushuhui/ares-contrib/middleware"
)

func main() {
    app := ares.New()

    // 使用默认配置的简单用法
    app.Use(middleware.RequestID())

    // 自定义配置
    app.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
        Generator: func() string {
            return "custom-id-" + uuid.New().String()
        },
        RequestIDHeader: "X-Request-ID",
        ContextKey:      "requestID",
    }))

    app.Run(":8080")
}
```

### Body Limit

限制最大请求体大小。

**特性：**
- 可配置的大小限制
- 使用 http.MaxBytesReader 进行高效限制
- 防止内存耗尽攻击

**使用方法：**

```go
import (
    "github.com/xushuhui/ares"
    "github.com/xushuhui/ares-contrib/middleware"
)

func main() {
    app := ares.New()

    // 限制请求体为 10MB
    app.Use(middleware.BodyLimit(10 * 1024 * 1024))

    app.Run(":8080")
}
```

### JWT 认证

基于令牌的身份验证 JWT 中间件。

**特性：**
- 使用 jwt.Keyfunc 灵活管理密钥
- 可配置的签名方法
- 自定义声明支持
- 详细的错误分类（过期、无效、格式错误）
- 基于上下文的声明存储

**使用方法：**

```go
import (
    "github.com/golang-jwt/jwt/v5"
    "github.com/xushuhui/ares"
    "github.com/xushuhui/ares-contrib/middleware"
)

func main() {
    app := ares.New()

    // 仅使用签名密钥的简单用法
    app.Use(middleware.JWT([]byte("your-secret-key")))

    // 使用自定义签名方法
    app.Use(middleware.JWT(
        []byte("your-secret-key"),
        middleware.WithSigningMethod(jwt.SigningMethodHS256),
    ))

    // 使用自定义声明
    app.Use(middleware.JWT(
        []byte("your-secret-key"),
        middleware.WithClaims(func() jwt.Claims {
            return &jwt.RegisteredClaims{}
        }),
        middleware.WithContextKey("user"),
    ))

    // 在处理器中访问声明
    app.GET("/protected", func(ctx *ares.Context) error {
        claims, ok := middleware.GetClaims(ctx.Request.Context())
        if !ok {
            return ctx.JSON(401, map[string]string{"error": "no claims"})
        }
        return ctx.JSON(200, claims)
    })

    app.Run(":8080")
}
```

**选项：**
- `WithSigningMethod(method)` - 设置 JWT 签名方法（默认：HS256）
- `WithClaims(func)` - 使用自定义声明结构
- `WithContextKey(key)` - 设置存储声明的自定义上下文键（默认："user"）

### 限流器

防止 API 滥用的限流中间件。

**特性：**
- 基于 IP 的限流
- 可配置的速率和突发
- 自定义键提取
- 自动清理旧的限流器

**使用方法：**

```go
import (
    "github.com/xushuhui/ares"
    "github.com/xushuhui/ares-contrib/middleware"
)

func main() {
    app := ares.New()

    // 使用默认配置的简单用法（10 req/s，突发 20）
    app.Use(middleware.RateLimiter())

    // 自定义配置
    app.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
        Rate:  100,  // 每秒 100 个请求
        Burst: 200,  // 允许突发 200 个请求
        KeyFunc: func(r *http.Request) string {
            // 自定义键提取（例如，按用户 ID）
            return r.Header.Get("X-User-ID")
        },
    }))

    app.Run(":8080")
}
```

### Gzip 压缩

减少带宽使用的响应压缩中间件。

**特性：**
- 可配置的压缩级别
- 最小响应大小阈值
- 排除特定文件扩展名
- 排除特定路径
- Writer 池化以提高性能

**使用方法：**

```go
import (
    "github.com/xushuhui/ares"
    "github.com/xushuhui/ares-contrib/middleware"
)

func main() {
    app := ares.New()

    // 使用默认配置的简单用法
    app.Use(middleware.Gzip())

    // 自定义配置
    app.Use(middleware.GzipWithConfig(middleware.GzipConfig{
        Level:     5,     // 压缩级别（1-9）
        MinLength: 1024,  // 仅压缩 > 1KB 的响应
        ExcludedExtensions: []string{".png", ".jpg", ".gif"},
        ExcludedPaths:      []string{"/api/stream"},
    }))

    app.Run(":8080")
}
```

**默认排除的扩展名：**
- 图片：`.png`、`.jpg`、`.jpeg`、`.gif`、`.webp`、`.svg`
- 压缩包：`.zip`、`.gz`、`.tar`、`.rar`、`.7z`
- 媒体：`.mp4`、`.avi`、`.mov`、`.mp3`、`.wav`
- 文档：`.pdf`

### 安全头

安全头中间件，用于防护常见的 Web 漏洞。

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
import (
    "github.com/xushuhui/ares"
    "github.com/xushuhui/ares-contrib/middleware/secure"
)

func main() {
    app := ares.New()

    // 使用默认配置
    app.Use(secure.New())

    // 自定义配置
    app.Use(secure.New(
        secure.WithXFrameOptions("DENY"),
        secure.WithHSTSMaxAge(31536000),  // 1 年
        secure.WithContentSecurityPolicy("default-src 'self'; script-src 'self' 'unsafe-inline'"),
        secure.WithReferrerPolicy("strict-origin-when-cross-origin"),
        secure.WithPermissionsPolicy("geolocation=(self), microphone=()"),
    ))

    app.Run(":8080")
}
```

**选项：**
- `WithXSSProtection(value)` - 设置 X-XSS-Protection 头（默认："1; mode=block"）
- `WithContentTypeNosniff(value)` - 设置 X-Content-Type-Options 头（默认："nosniff"）
- `WithXFrameOptions(value)` - 设置 X-Frame-Options 头（默认："SAMEORIGIN"）
- `WithHSTSMaxAge(seconds)` - 设置 HSTS max-age（秒）（默认：0，禁用）
- `WithHSTSExcludeSubdomains(bool)` - 从 HSTS 中排除子域（默认：false）
- `WithContentSecurityPolicy(policy)` - 设置 Content-Security-Policy 头
- `WithCSPReportOnly(bool)` - 使用 CSP 仅报告模式（默认：false）
- `WithReferrerPolicy(policy)` - 设置 Referrer-Policy 头
- `WithPermissionsPolicy(policy)` - 设置 Permissions-Policy 头

**默认头：**
- `X-XSS-Protection: 1; mode=block`
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: SAMEORIGIN`

## 依赖

- `github.com/golang-jwt/jwt/v5` - JWT 实现
- `github.com/google/uuid` - UUID 生成
- `golang.org/x/time/rate` - 限流

## 未来计划

此 contrib 包最终将移至单独的仓库，以允许独立的版本控制和开发。

## 许可证

MIT
