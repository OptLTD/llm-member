# Web / Wails 客户端集成

## 登录（浏览器 OAuth + scheme 回调）

1. Wails WebView 打开 `{MEMBER_URL}/signin`
2. 用户登录后调用 `/v1/build-token`
3. 服务端重定向到 `AUTH_CALLBACK?token=&sign=&time=`
4. Wails Go 壳捕获 scheme，导航 WebView 到 `/auth-callback?...`
5. WebView JS 调用 `POST /v1/verify-token` 获取 `userPlan`、`expireAt`、`apiToken`

## 设备绑定

```javascript
const deviceId = localStorage.getItem('member_device_id') || crypto.randomUUID();
localStorage.setItem('member_device_id', deviceId);
await fetch('/api/device/bind', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json' 
  },
  body: JSON.stringify({ deviceId, platform: navigator.platform, appVersion: '1.0.0' }),
});
```

## 埋点 SDK

单实例部署无需 `siteKey`，服务端自动归入默认站点：

```html
<script src="https://your-member.com/sdk/analytics.js"></script>
```

- 自动：`app.launch`、`app.page_view`、`session.start/tick/end`
- 主动：`MemberTrack('feature.export', { format: 'pdf' })`

## 券码

- 下单前：`POST /api/coupon/validate`
- 延期券兑换：`POST /api/coupon/redeem`

## 问卷 / 反馈

- 问卷：`GET /api/survey/:id`，`POST /api/survey/:id/submit`
- 反馈：`POST /api/feedback` 或打开 `/feedback` 页面
