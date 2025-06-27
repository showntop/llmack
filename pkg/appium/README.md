# Appium Go Client

Appium Go客户端，从Python Appium客户端改写而来，使用[github.com/tebeka/selenium]作为底层WebDriver实现。

## 特性

- 🚀 **完整的Appium功能** - 支持所有主要的Appium功能和命令
- 🔧 **扩展系统** - 支持自定义扩展，类似Python版本的ExtensionBase
- 📱 **多平台支持** - 支持Android、iOS等平台
- ⚡ **高性能** - 基于Go语言的高性能实现
- 🛡️ **类型安全** - 完全的类型安全，减少运行时错误
- 🔗 **直连支持** - 支持Appium的直连功能

## 安装

```bash
go mod init your-project
go get github.com/tebeka/selenium
go get github.com/google/uuid
```

## 快速开始

### 基本用法

```go
package main

import (
    "fmt"
    "time"
    appium "path/to/appium-go"
)

func main() {
    // 创建Appium选项
    options := appium.NewAppiumOptions().
        SetPlatformName("Android").
        SetDeviceName("emulator-5554").
        SetApp("/path/to/your/app.apk").
        SetAutomationName("UIAutomator2")
    
    // 创建客户端配置
    clientConfig := appium.NewAppiumClientConfig("http://127.0.0.1:4723").
        WithTimeout(30 * time.Second).
        WithKeepAlive(true)
    
    // 创建WebDriver
    driver, err := appium.NewWebDriver("", options, clientConfig, nil)
    if err != nil {
        panic(err)
    }
    defer driver.Close()
    
    // 启动会话
    err = driver.StartSession(options.ToCapabilities())
    if err != nil {
        panic(err)
    }
    
    // 获取服务器状态
    status, err := driver.GetStatus()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("服务器状态: %+v\n", status)
}
```

### 使用Applications扩展

```go
package main

import (
    appium "path/to/appium-go"
)

func main() {
    // 创建driver（省略配置代码）
    options := appium.NewAppiumOptions().
        SetPlatformName("Android").
        SetDeviceName("emulator-5554").
        SetAutomationName("UIAutomator2")
    
    driver, err := appium.NewWebDriver("", options, nil, nil)
    if err != nil {
        panic(err)
    }
    defer driver.Close()
    
    // 创建Applications扩展
    apps := appium.NewApplications(driver)
    
    bundleID := "com.example.app"
    
    // 检查应用是否已安装
    installed, err := apps.IsAppInstalled(bundleID)
    if err != nil {
        panic(err)
    }
    
    if !installed {
        // 安装应用
        err = apps.InstallApp("/path/to/app.apk", map[string]interface{}{
            "replace": true,
        })
        if err != nil {
            panic(err)
        }
    }
    
    // 激活应用
    err = apps.ActivateApp(bundleID)
    if err != nil {
        panic(err)
    }
    
    // 将应用置于后台5秒
    err = apps.BackgroundApp(5)
    if err != nil {
        panic(err)
    }
    
    // 查询应用状态
    state, err := apps.QueryAppState(bundleID)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("应用状态: %d\n", state)
}
```

### 创建自定义扩展

```go
package main

import (
    appium "path/to/appium-go"
)

// 自定义扩展示例
type MyCustomExtension struct {
    appium.BaseExtension
}

func (ext *MyCustomExtension) MethodName() string {
    return "myCustomMethod"
}

func (ext *MyCustomExtension) AddCommand() (string, string) {
    return "GET", "session/$sessionId/my/custom/endpoint"
}

func (ext *MyCustomExtension) MyCustomMethod() (interface{}, error) {
    result, err := ext.Execute(nil)
    if err != nil {
        return nil, err
    }
    
    // 提取value字段（Appium响应格式）
    if resultMap, ok := result.(map[string]interface{}); ok {
        if value, exists := resultMap["value"]; exists {
            return value, nil
        }
    }
    
    return result, nil
}

func main() {
    // 创建自定义扩展
    customExt := &MyCustomExtension{}
    
    // 创建带有自定义扩展的driver
    options := appium.NewAppiumOptions().
        SetPlatformName("iOS").
        SetDeviceName("iPhone Simulator").
        SetAutomationName("XCUITest")
    
    driver, err := appium.NewWebDriver("", options, nil, []appium.ExtensionBase{customExt})
    if err != nil {
        panic(err)
    }
    defer driver.Close()
    
    // 使用自定义扩展方法
    result, err := customExt.MyCustomMethod()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("自定义扩展结果: %v\n", result)
}
```

### 使用直连功能

```go
package main

import (
    appium "path/to/appium-go"
)

func main() {
    // 启用直连功能的客户端配置
    clientConfig := appium.NewAppiumClientConfig("http://127.0.0.1:4723").
        WithDirectConnection(true).
        WithKeepAlive(true)
    
    options := appium.NewAppiumOptions().
        SetPlatformName("Android").
        SetDeviceName("emulator-5554").
        SetAutomationName("UIAutomator2")
    
    driver, err := appium.NewWebDriver("", options, clientConfig, nil)
    if err != nil {
        panic(err)
    }
    defer driver.Close()
    
    // 启动会话 - 如果服务器支持，将使用直连
    err = driver.StartSession(options.ToCapabilities())
    if err != nil {
        panic(err)
    }
    
    fmt.Println("会话已启动，使用直连功能")
}
```

## 核心概念

### AppiumOptions

`AppiumOptions` 类用于构建Appium capabilities，支持链式调用：

```go
options := appium.NewAppiumOptions().
    SetPlatformName("Android").
    SetDeviceName("emulator-5554").
    SetApp("/path/to/app.apk").
    SetAutomationName("UIAutomator2").
    SetUDID("device-udid").
    SetNoReset(true).
    SetFullReset(false).
    SetNewCommandTimeout(60)
```

### AppiumClientConfig

`AppiumClientConfig` 用于配置客户端行为：

```go
clientConfig := appium.NewAppiumClientConfig("http://127.0.0.1:4723").
    WithDirectConnection(true).
    WithKeepAlive(true).
    WithTimeout(30 * time.Second).
    WithCapability("customCap", "value")
```

### 扩展系统

扩展系统允许你添加自定义功能，类似Python版本：

1. 实现 `ExtensionBase` 接口
2. 提供 `MethodName()` 和 `AddCommand()` 方法
3. 创建WebDriver时传入扩展

### 命令执行

所有Appium命令都通过 `Execute` 方法执行：

```go
result, err := driver.Execute("getStatus", nil)
```

## API 参考

### WebDriver 方法

- `NewWebDriver(commandExecutor, options, clientConfig, extensions)` - 创建新的WebDriver实例
- `StartSession(capabilities)` - 启动新会话
- `GetStatus()` - 获取服务器状态
- `GetOrientation()` - 获取设备方向
- `SetOrientation(orientation)` - 设置设备方向
- `Execute(command, params)` - 执行命令
- `Close()` - 关闭会话

### Applications 方法

- `IsAppInstalled(bundleID)` - 检查应用是否已安装
- `InstallApp(appPath, options)` - 安装应用
- `RemoveApp(appID, options)` - 卸载应用
- `ActivateApp(appID)` - 激活应用
- `TerminateApp(appID, options)` - 终止应用
- `BackgroundApp(seconds)` - 将应用置于后台
- `QueryAppState(appID)` - 查询应用状态
- `GetAppStrings(language, stringFile)` - 获取应用字符串

## 与Python版本的对比

| 特性 | Python | Go |
|------|--------|----| 
| 类型安全 | 运行时检查 | 编译时检查 |
| 性能 | 解释执行 | 编译执行 |
| 扩展系统 | 动态添加方法 | 接口实现 |
| 错误处理 | 异常 | error返回值 |
| 并发 | 线程/协程 | goroutine |

## 注意事项

1. 这是一个基础实现，某些高级功能可能需要进一步开发
2. 需要配合Appium服务器使用
3. 确保设备连接和配置正确
4. 建议在测试环境中充分验证

## 贡献

欢迎提交PR和Issue！

## 许可证

Apache License 2.0 