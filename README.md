# talkFlow（语流）

一款开黑语音工具，用户只要打开 web 界面，就能直接和好友进行语言通话进行开黑

## 特点

- 便捷：不需要安装额外的桌面程序或者APP，直接打开浏览器即可开黑
- 轻量：后端使用 Go 编写，只要是台服务器就能跑
- 傻瓜式：提供 Docker 部署方式，方便部署和迁移

## 技术栈

### 后端

- Gin
- MongoDB：感觉可以提供可选项，如果用户没有设置则使用内置的SQLite

## 设计思路

参考 FileCodeBox 的设计思路：

> 它允许用户通过简单的方式分享文本和文件，接收者只需要一个提取码就可以取得文件，就像从快递柜取出快递一样简单

talkFlow 同样可以允许用户通过简单的方式来进行开黑，只需要用户生成一个邀请码，然后在主页里面输入，即可进入群聊，邀请码具有时效性，而且管理员可以设置谁可以生成验证码，确保私域性

## 开发

需要自行在本地配置好 MongoDB 环境，并且设置好 `JWT_SECRET`

```bash
cp .env.example .env
openssl rand -hex 32
go run main.go
```

## 后端

### 用户相关

```
# 用户认证
POST   /api/v1/auth/register  { username, password }
POST   /api/v1/auth/login     { username, password } → { token }

# 需带上 Auth 鉴权
GET    /api/v1/profile
```