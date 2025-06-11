// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	appiumgo "appium-go"
)

func main() {
	fmt.Println("🚀 Appium Go客户端演示程序")
	fmt.Println("================================")

	// 演示1: 基本用法
	fmt.Println("\n📱 演示1: 基本WebDriver功能")
	demoBasicUsage()
	// 演示2: 应用管理
	fmt.Println("\n📦 演示2: 应用管理功能")
	demoApplicationManagement()

	// 演示3: 选项构建器
	fmt.Println("\n⚙️ 演示3: 选项构建器")
	demoOptionsBuilder()

	// 演示4: W3C capabilities转换
	fmt.Println("\n🌐 演示4: W3C Capabilities转换")
	demoW3CConversion()

	fmt.Println("\n✅ 所有演示完成！")
}

// demoBasicUsage 演示基本WebDriver功能
func demoBasicUsage() {
	// "appium:automationName":    "UiAutomator2",
	// "platformName":             "Android",
	// "appium:udid":              "emulator-5554",
	// "appium:deviceName":        "emulator-5554",
	// "appium:appPackage":        "com.android.settings",
	// "appium:appActivity":       ".Settings",
	// "appium:noReset":           true,
	// "appium:newCommandTimeout": 3600,
	fmt.Println("创建Appium选项...")
	options := appiumgo.NewAppiumOptions().
		SetPlatformName("Android").
		SetDeviceName("emulator-5554").
		// SetApp("/path/to/your/app.apk").
		SetAutomationName("UIAutomator2").
		SetNoReset(true).
		SetNewCommandTimeout(3600)

	fmt.Println("创建客户端配置...")
	clientConfig := appiumgo.NewAppiumClientConfig("http://127.0.0.1:4723").
		WithTimeout(30 * time.Second).
		WithKeepAlive(true)

	fmt.Println("创建WebDriver实例...")
	driver, err := appiumgo.NewWebDriver("", options, clientConfig, nil)
	if err != nil {
		log.Printf("❌ 创建WebDriver失败: %v", err)
		return
	}
	defer func() {
		if err := driver.Close(); err != nil {
			log.Printf("❌ 关闭driver失败: %v", err)
		}
	}()

	fmt.Println("✅ WebDriver创建成功")

	screenshot, err := driver.Screenshot()
	if err != nil {
		log.Printf("❌ 截图失败: %v", err)
	} else {
		os.WriteFile("screenshotx.png", screenshot, 0644)
		log.Printf("✅ 截图成功: %v", len(screenshot))
	}

	if err := driver.Tap([]appiumgo.Position{{X: 166, Y: 1983}}, 1); err != nil {
		log.Printf("❌ 点击失败: %v", err)
	}

	return

	// 模拟启动会话（在实际环境中需要Appium服务器）
	fmt.Println("尝试启动会话...")
	err = driver.StartSession(options.ToCapabilities())
	if err != nil {
		log.Printf("⚠️ 启动会话失败（需要Appium服务器）: %v", err)
	} else {
		fmt.Println("✅ 会话启动成功")

		// 获取服务器状态
		fmt.Println("获取服务器状态...")
		status, err := driver.GetStatus()
		if err != nil {
			log.Printf("❌ 获取状态失败: %v", err)
		} else {
			fmt.Printf("✅ 服务器状态: %+v\n", status)
		}

		// 获取设备方向
		fmt.Println("获取设备方向...")
		orientation, err := driver.GetOrientation()
		if err != nil {
			log.Printf("❌ 获取方向失败: %v", err)
		} else {
			fmt.Printf("✅ 当前方向: %s\n", orientation)
		}
	}
}

// demoApplicationManagement 演示应用管理功能
func demoApplicationManagement() {
	fmt.Println("创建WebDriver和Applications扩展...")

	options := appiumgo.NewAppiumOptions().
		SetPlatformName("Android").
		SetDeviceName("emulator-5554").
		SetAutomationName("UIAutomator2")

	driver, err := appiumgo.NewWebDriver("", options, nil, nil)
	if err != nil {
		log.Printf("❌ 创建WebDriver失败: %v", err)
		return
	}
	defer driver.Close()

	// 创建Applications扩展
	apps := appiumgo.NewApplications(driver)
	bundleID := "com.example.testapp"

	fmt.Printf("检查应用是否已安装: %s\n", bundleID)
	installed, err := apps.IsAppInstalled(bundleID)
	if err != nil {
		log.Printf("⚠️ 检查应用安装状态失败（需要活跃会话）: %v", err)
	} else {
		if installed {
			fmt.Println("✅ 应用已安装")
		} else {
			fmt.Println("ℹ️ 应用未安装")
		}
	}

	// 演示其他应用管理功能
	fmt.Println("演示应用管理API调用...")

	// 安装应用
	fmt.Println("- 安装应用API")
	err = apps.InstallApp("/path/to/app.apk", map[string]interface{}{
		"replace": true,
		"timeout": 60000,
	})
	if err != nil {
		log.Printf("⚠️ 安装应用失败（需要活跃会话）: %v", err)
	}

	// 激活应用
	fmt.Println("- 激活应用API")
	err = apps.ActivateApp(bundleID)
	if err != nil {
		log.Printf("⚠️ 激活应用失败（需要活跃会话）: %v", err)
	}

	// 查询应用状态
	fmt.Println("- 查询应用状态API")
	state, err := apps.QueryAppState(bundleID)
	if err != nil {
		log.Printf("⚠️ 查询应用状态失败（需要活跃会话）: %v", err)
	} else {
		fmt.Printf("✅ 应用状态: %d\n", state)
	}

	// 将应用置于后台
	fmt.Println("- 后台应用API")
	err = apps.BackgroundApp(5)
	if err != nil {
		log.Printf("⚠️ 后台应用失败（需要活跃会话）: %v", err)
	}

	// 终止应用
	fmt.Println("- 终止应用API")
	terminated, err := apps.TerminateApp(bundleID, nil)
	if err != nil {
		log.Printf("⚠️ 终止应用失败（需要活跃会话）: %v", err)
	} else {
		fmt.Printf("✅ 应用终止结果: %t\n", terminated)
	}

	fmt.Println("✅ 应用管理功能演示完成")
}

// demoOptionsBuilder 演示选项构建器
func demoOptionsBuilder() {
	fmt.Println("演示选项构建器的链式调用...")

	// Android选项
	androidOptions := appiumgo.NewAppiumOptions().
		SetPlatformName("Android").
		SetDeviceName("Pixel_4_API_30").
		SetApp("/path/to/android/app.apk").
		SetAutomationName("UIAutomator2").
		SetUDID("emulator-5554").
		SetNoReset(true).
		SetFullReset(false).
		SetNewCommandTimeout(120)

	fmt.Println("✅ Android选项构建完成")
	androidCaps := androidOptions.ToCapabilities()
	fmt.Printf("Android Capabilities数量: %d\n", len(androidCaps))

	// iOS选项
	iOSOptions := appiumgo.NewAppiumOptions().
		SetPlatformName("iOS").
		SetDeviceName("iPhone 13").
		SetApp("/path/to/ios/app.app").
		SetAutomationName("XCUITest").
		SetUDID("00008030-001C2D8E3A82802E").
		SetNoReset(false).
		SetNewCommandTimeout(90)

	fmt.Println("✅ iOS选项构建完成")
	iOSCaps := iOSOptions.ToCapabilities()
	fmt.Printf("iOS Capabilities数量: %d\n", len(iOSCaps))

	// 自定义capabilities
	customOptions := appiumgo.NewAppiumOptions().
		SetCapability("customCap1", "value1").
		SetCapability("customCap2", 123).
		SetCapability("customCap3", true).
		LoadCapabilities(map[string]interface{}{
			"batchCap1": "batchValue1",
			"batchCap2": 456,
			"batchCap3": false,
		})

	fmt.Println("✅ 自定义选项构建完成")
	customCaps := customOptions.ToCapabilities()
	fmt.Printf("自定义Capabilities数量: %d\n", len(customCaps))

	// 获取特定capability
	fmt.Println("获取特定capability值...")
	platformName := androidOptions.GetCapability("platformName")
	deviceName := androidOptions.GetCapability("deviceName")
	nonExistent := androidOptions.GetCapability("nonExistentCap")

	fmt.Printf("platformName: %v\n", platformName)
	fmt.Printf("deviceName: %v\n", deviceName)
	fmt.Printf("nonExistentCap: %v\n", nonExistent)
}

// demoW3CConversion 演示W3C capabilities转换
func demoW3CConversion() {
	fmt.Println("演示W3C capabilities转换...")

	// 原始capabilities（包含OSS格式）
	originalCaps := map[string]interface{}{
		"platformName":     "iOS",
		"deviceName":       "iPhone 12",
		"automationName":   "XCUITest",
		"acceptSslCerts":   true,   // OSS格式，应转换为acceptInsecureCerts
		"version":          "14.5", // OSS格式，应转换为browserVersion
		"platform":         "iOS",  // OSS格式，应转换为platformName
		"customCapability": "customValue",
		"app":              "/path/to/app.app",
	}

	fmt.Printf("原始capabilities数量: %d\n", len(originalCaps))

	// 转换为W3C格式
	w3cCaps := appiumgo.AsW3C(originalCaps)
	fmt.Println("✅ W3C转换完成")

	// 检查转换结果
	if capabilities, exists := w3cCaps["capabilities"]; exists {
		caps := capabilities.(map[string]interface{})

		if alwaysMatch, exists := caps["alwaysMatch"]; exists {
			alwaysMatchCaps := alwaysMatch.(map[string]interface{})
			fmt.Printf("alwaysMatch capabilities数量: %d\n", len(alwaysMatchCaps))

			// 检查OSS到W3C的转换
			if acceptInsecureCerts, exists := alwaysMatchCaps["acceptInsecureCerts"]; exists {
				fmt.Printf("✅ acceptSslCerts -> acceptInsecureCerts: %v\n", acceptInsecureCerts)
			}

			if browserVersion, exists := alwaysMatchCaps["browserVersion"]; exists {
				fmt.Printf("✅ version -> browserVersion: %v\n", browserVersion)
			}

			// 检查appium前缀
			if deviceName, exists := alwaysMatchCaps["appium:deviceName"]; exists {
				fmt.Printf("✅ deviceName -> appium:deviceName: %v\n", deviceName)
			}

			if customCap, exists := alwaysMatchCaps["appium:customCapability"]; exists {
				fmt.Printf("✅ customCapability -> appium:customCapability: %v\n", customCap)
			}
		}

		if firstMatch, exists := caps["firstMatch"]; exists {
			firstMatchArray := firstMatch.([]map[string]interface{})
			fmt.Printf("firstMatch数组长度: %d\n", len(firstMatchArray))
		}
	}

	// 使用AppiumOptions进行转换
	fmt.Println("\n使用AppiumOptions进行W3C转换...")
	options := appiumgo.NewAppiumOptions().
		LoadCapabilities(originalCaps)

	optionsW3C := options.ToW3C()
	fmt.Println("✅ AppiumOptions W3C转换完成")

	if capabilities, exists := optionsW3C["capabilities"]; exists {
		caps := capabilities.(map[string]interface{})
		if alwaysMatch, exists := caps["alwaysMatch"]; exists {
			alwaysMatchCaps := alwaysMatch.(map[string]interface{})
			fmt.Printf("AppiumOptions W3C capabilities数量: %d\n", len(alwaysMatchCaps))
		}
	}
}
