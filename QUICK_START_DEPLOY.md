# Quick Start - Docker Deployment

## 🚀 Deploy in 3 Steps

### 1. Set Up Environment
Create `.env` file:
```env
SERVER_PORT=8080
DATABASE_URL=postgres://user:password@neon-host/dbname?sslmode=require
DB_MAX_CONNECTIONS=10
```

### 2. Build & Run
```bash
docker-compose up --build
```

### 3. Access
Open http://localhost:8080

## 📦 What Was Added

- ✅ `Dockerfile` - Multi-stage build for production
- ✅ `docker-compose.yml` - Easy local/production deployment
- ✅ `.dockerignore` - Optimized builds
- ✅ Database connection pooling - Now using MaxConnections properly
- ✅ `DEPLOYMENT.md` - Full deployment guide
- ✅ `NEON_DB_ANALYSIS.md` - Neon DB optimization guide

## 🔧 Quick Commands

```bash
# Build image
docker build -t rental-app .

# Run locally
docker run -p 8080:8080 --env-file .env rental-app

# View logs (if running in background)
docker logs -f rental-app

# Stop container
docker stop rental-app
```

## 🌐 Deploy to Cloud (FREE Options)

**🎯 Recommended: Render.com (No Credit Card)**
1. Sign up at [render.com](https://render.com) (free, no credit card)
2. Connect GitHub repo
3. Add `DATABASE_URL` environment variable
4. Deploy (auto-detects Dockerfile)
5. Done! Your app: `https://your-app.onrender.com`

**Other Free Options:**
- **Fly.io** - Always-on, no credit card
- **Cyclic.sh** - Simple, free forever
- **Railway** - Easy setup (may ask card but won't charge)

📖 **See [FREE_DEPLOYMENT.md](./FREE_DEPLOYMENT.md) for complete free deployment guide!**

## 💡 Neon DB Tip

For **free tier**, set `DB_MAX_CONNECTIONS=10` in your `.env` file to avoid connection limit issues.

