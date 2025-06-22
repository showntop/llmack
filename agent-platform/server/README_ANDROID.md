# Android Agent 支持说明

## 功能概述

服务器现在支持 `@android` 功能，可以通过智能代理控制 Android 设备执行各种任务。

## 支持的任务类型

### 1. 购物类任务
- 打开购物应用（淘宝、京东、小红书等）
- 搜索商品
- 浏览商品详情和用户评价
- 比较不同商品
- 提供购买建议

### 2. 应用操作
- 启动应用
- 点击界面元素
- 输入文本
- 滑动操作
- 截屏

### 3. 设备管理
- 安装应用
- 获取设备状态
- 列出已安装应用

## 使用方法

### 触发 Android Agent

在消息中包含以下任何关键词会自动触发 Android Agent：

- `@android`
- `android`
- `手机`
- `mobile`
- `app`
- `应用`
- `点击`
- `滑动`
- `截屏`
- `安装`
- `启动应用`
- `手机操作`
- 应用名称：`淘宝`、`京东`、`小红书`、`微信`、`抖音`、`支付宝`
- 购物相关：`购物`、`下单`、`浏览`、`搜索商品`、`打开应用`

### 示例任务

```
我想购买一辆儿童自行车供3岁小孩使用。
1. 仔细思考需求，并列出需要考虑的点，综合考虑后再决策。
2. 你可以浏览器及小红书、淘宝、京东等APP，帮助我决策购物。
3. 你是一个经验丫富的购物专家，可以帮我分析商品，给出建议。
4. 你尤其需要关注真实用户的评论。
```

```
帮我打开抖音应用，搜索关于编程的视频
```

```
@android 截屏当前界面，然后分析屏幕上的内容
```

## Android 设备要求

1. **设备连接**: 确保 Android 设备通过 ADB 连接到服务器
2. **必要应用**: 需要安装以下应用：
   - `droidrun-portable.apk` - 设备控制服务
   - `ADBKeyboard.apk` - 文本输入支持

3. **设备配置**:
   ```bash
   adb install droidrun-portable.apk
   adb install ADBKeyboard.apk
   adb shell ime enable com.android.adbkeyboard/.AdbIME
   adb shell ime set com.android.adbkeyboard/.AdbIME
   ```

## 环境配置

创建 `.env` 文件：
```
doubao_api_key=your_doubao_api_key_here
claude_api_key=your_claude_api_key_here
qwen_api_key=your_qwen_api_key_here
ANDROID_DEVICE_SERIAL=47.96.179.122:1000
```

## API 响应格式

### Android 任务步骤
当检测到 Android 任务时，系统会返回以下执行步骤：

1. 📱 初始化Android设备连接
2. 📋 分析任务需求
3. 🔍 获取当前屏幕状态
4. 🎯 定位目标元素
5. ⚡ 执行Android操作
6. ✅ 验证操作结果
7. 📤 生成任务报告

### 会话信息
响应中会包含 `agent_type` 字段标识使用的代理类型：
- `"ai"` - 标准 AI 助手
- `"android"` - Android 移动设备代理

## 注意事项

1. **设备准备**: 确保 Android 设备已连接且可以通过 ADB 访问
2. **权限设置**: 某些操作可能需要特定的系统权限
3. **网络连接**: 设备需要稳定的网络连接
4. **响应时间**: Android 操作可能需要较长时间，请耐心等待
5. **错误处理**: 如果设备连接失败，系统会自动切换到标准 AI 助手模式

## 故障排除

### 常见问题

1. **设备未连接**
   - 检查 ADB 连接状态：`adb devices`
   - 确认设备序列号配置正确

2. **应用启动失败**
   - 确认目标应用已安装
   - 检查应用权限设置

3. **UI 元素定位失败**
   - 等待界面完全加载
   - 检查屏幕截图确认当前状态

4. **文本输入问题**
   - 确认 ADBKeyboard 已正确安装和配置
   - 检查输入法设置 