# talkFlow（语流）

> 该项目目前正处于活跃开发状态，请不要将其用于生产环境中！

一款开黑语音工具，用户只要打开 web 界面，就能直接和好友进行语言通话进行开黑

## 特点

- 便捷：不需要安装额外的桌面程序或者APP，直接打开浏览器即可开黑
- 轻量：后端使用 Go 编写，只要是台服务器就能跑
- 傻瓜式：提供 Docker 部署方式，方便部署和迁移

## 技术栈

### 后端

- Gin

## 设计思路

参考 FileCodeBox 的设计思路：

> 它允许用户通过简单的方式分享文本和文件，接收者只需要一个提取码就可以取得文件，就像从快递柜取出快递一样简单

talkFlow 同样可以允许用户通过简单的方式来进行开黑，只需要用户生成一个邀请码，然后在主页里面输入，即可进入群聊，邀请码具有时效性，而且管理员可以设置谁可以生成验证码，确保私域性

## 开发

```bash
cp .env.example .env
openssl rand -hex 32
go run main.go
```

## 后端

后端的 API 文档托管在：https://talkflow.apifox.cn

目前已经提供了 Docker 镜像，如何要在本地测试的话可以直接使用：

```bash
docker compose up -d
```

### 用户相关

```
# 用户认证
POST   /api/v1/auth/register  { username, password }
POST   /api/v1/auth/login     { username, password } → { token }

# 需带上 Auth 鉴权
GET    /api/v1/profile
```

### 聊天相关

```
# 创建房间（Box）
POST /api/v1/room/create { name, expire_time } → { code }
POST /api/v1/room/join   { join_code, visitor_id } → { url }
GET  /api/v1/ws          { join_code, visitor_id }
```

## 开发日志

- [x] 用户的注册和登录
- [x] 聊天频道的创建（Box）
- [x] 从 MongoDB 迁移到 SQLite
- [ ] WebSocket 实现