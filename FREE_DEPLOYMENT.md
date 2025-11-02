# Free Docker Deployment (No Credit Card Required)

Complete guide for deploying your rental management app using **100% free** services that **don't require credit cards**.

## ✅ Best Options (Ranked)

### 1. 🥇 Render.com (RECOMMENDED)

**Why:** Most reliable free tier, easy setup, no credit card needed

**Free Tier Limits:**
- ✅ 750 free instance-hours/month per workspace
- ✅ Free SSL certificates
- ✅ Custom domains
- ⚠️ Services spin down after 15 min inactivity (wakes up in ~30 seconds)
- ⚠️ 512 MB RAM limit

**Deployment Steps:**

1. **Sign up** at [render.com](https://render.com) (no credit card needed)

2. **Create New Web Service:**
   - Click "New +" → "Web Service"
   - Connect your GitHub repository
   - Select the repository: `pythonProject2`

3. **Configure:**
   ```
   Name: rental-management-app
   Region: Choose closest to you
   Branch: main (or master)
   Root Directory: (leave empty)
   Environment: Docker
   Dockerfile Path: Dockerfile  (this file already exists in your repo ✅)
   Docker Context: ./
   ```
   
   **Note:** Render only needs `Dockerfile` - no docker-compose needed!

4. **Environment Variables:**
   Click "Environment" tab and add:
   ```
   SERVER_PORT=8080
   DATABASE_URL=your-neon-db-url
   DB_SSL_MODE=require
   DB_MAX_CONNECTIONS=10
   LOG_LEVEL=info
   ```

5. **Deploy:**
   - Click "Create Web Service"
   - Wait for build to complete (~5-10 minutes)
   - Your app will be live at: `https://rental-management-app.onrender.com`

**Pros:**
- ✅ No credit card required
- ✅ Easy GitHub integration
- ✅ Auto-deploys on git push
- ✅ Free SSL included
- ✅ Reliable uptime

**Cons:**
- ⚠️ Spins down after 15 min inactivity (first request may take 30-60 seconds)
- ⚠️ Limited RAM (but sufficient for your app)

---

### 2. 🥈 Fly.io

**Why:** Generous free tier, always-on, great performance

**Free Tier Limits:**
- ✅ 3 shared-cpu-1x VMs (256 MB RAM each)
- ✅ 160 GB outbound data transfer/month
- ✅ Always-on (doesn't sleep)
- ⚠️ Requires CLI setup (slightly more complex)

**Deployment Steps:**

1. **Install Fly CLI:**
   ```bash
   # Mac/Linux
   curl -L https://fly.io/install.sh | sh
   
   # Or download from: https://fly.io/docs/hands-on/install-flyctl/
   ```

2. **Sign up:**
   ```bash
   fly auth signup
   # No credit card required
   ```

3. **Create fly.toml:**
   Create `fly.toml` in project root:
   ```toml
   app = "rental-management-app"
   primary_region = "iad"  # Change to your region (iad = US East)

   [build]
     dockerfile = "Dockerfile"

   [http_service]
     internal_port = 8080
     force_https = true

     [[http_service.ports]]
       port = 80
       handlers = ["http"]
     [[http_service.ports]]
       port = 443
       handlers = ["tls", "http"]

   [env]
     SERVER_PORT = "8080"
     LOG_LEVEL = "info"
     DB_SSL_MODE = "require"
     DB_MAX_CONNECTIONS = "10"
   ```

4. **Deploy:**
   ```bash
   # Launch app
   fly launch
   
   # Set database URL (keep secret)
   fly secrets set DATABASE_URL="your-neon-db-url"
   
   # Open in browser
   fly open
   ```

**Pros:**
- ✅ Always-on (no cold starts)
- ✅ Generous free tier
- ✅ Great performance
- ✅ Multiple regions

**Cons:**
- ⚠️ Requires CLI knowledge
- ⚠️ Slightly more setup

---

### 3. 🥉 Railway.app

**Why:** Very easy setup, but requires credit card for verification (but won't charge on free tier)

**Note:** Railway **may** ask for credit card but won't charge on free tier

**Free Tier Limits:**
- ✅ $5 free credit/month
- ✅ ~500 hours free usage
- ✅ Auto-deploy from GitHub

**Deployment Steps:**

1. **Sign up** at [railway.app](https://railway.app)
   - May ask for credit card but won't charge on free tier

2. **New Project:**
   - Click "New Project"
   - Select "Deploy from GitHub repo"
   - Choose your repository

3. **Configure:**
   - Railway auto-detects Dockerfile
   - Add environment variables:
     ```
     DATABASE_URL=your-neon-db-url
     SERVER_PORT=8080
     DB_MAX_CONNECTIONS=10
     ```

4. **Deploy:**
   - Railway auto-deploys
   - Get your URL: `https://your-app.up.railway.app`

**Pros:**
- ✅ Easiest setup
- ✅ Auto-detects Docker
- ✅ Good documentation

**Cons:**
- ⚠️ May ask for credit card (but free tier is truly free)

---

### 4. Cyclic.sh

**Why:** Serverless, free forever, very simple

**Free Tier Limits:**
- ✅ Unlimited free apps
- ✅ No credit card required
- ✅ Auto-scales
- ⚠️ Sleeps after 5 min inactivity

**Deployment Steps:**

1. **Sign up** at [cyclic.sh](https://cyclic.sh)

2. **Connect GitHub:**
   - Click "Add App"
   - Select your repository
   - Cyclic auto-detects Dockerfile

3. **Environment Variables:**
   ```
   DATABASE_URL=your-neon-db-url
   SERVER_PORT=8080
   ```

4. **Deploy:**
   - Automatic deployment
   - Your app: `https://your-app.cyclic.app`

**Pros:**
- ✅ Completely free
- ✅ No credit card
- ✅ Very simple

**Cons:**
- ⚠️ Sleeps after inactivity
- ⚠️ Newer platform (less established)

---

## 🎯 Recommended Setup (Render.com)

For your rental management app, **Render.com** is the best choice:

### Complete Render Setup

1. **Create Account:**
   ```
   Go to: https://render.com
   Click: "Get Started for Free"
   Sign up with GitHub (recommended)
   ```

2. **Deploy Service:**
   ```
   Dashboard → New + → Web Service
   ```

3. **Configuration:**
   ```
   Name: rental-app
   Environment: Docker
   Region: Singapore (or closest)
   Branch: main
   ```

4. **Environment Variables:**
   ```env
   SERVER_PORT=8080
   DATABASE_URL=postgres://user:pass@neon-host/db?sslmode=require
   DB_SSL_MODE=require
   DB_MAX_CONNECTIONS=10
   LOG_LEVEL=info
   ```

5. **Advanced Settings:**
   ```
   Auto-Deploy: Yes
   Health Check Path: /login
   ```

6. **Deploy!**
   - Click "Create Web Service"
   - Build takes ~5-10 minutes
   - Your app: `https://rental-app.onrender.com`

### Handle Cold Starts (Render)

Since Render spins down after inactivity, the first request may be slow. Solutions:

**Option 1: Keep-Alive Script** (Recommended)
Create a free UptimeRobot account:
```
Service: HTTP(S)
URL: https://rental-app.onrender.com/login
Interval: 14 minutes (before 15 min timeout)
```

**Option 2: Accept Cold Starts**
- Free users: Just wait ~30 seconds for first request
- Fine for personal/small projects

---

## 🔄 Auto-Deploy Setup

### GitHub Actions (Optional)

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy to Render

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger Render Deploy
        run: |
          curl -X POST https://api.render.com/deploy/srv/YOUR_SERVICE_ID \
            -H "Authorization: Bearer ${{ secrets.RENDER_API_KEY }}"
```

---

## 📊 Comparison Table

| Service | Credit Card | Always-On | Free Tier | Difficulty |
|---------|-------------|-----------|-----------|------------|
| **Render** | ❌ No | ❌ Sleeps | 750 hrs/mo | ⭐ Easy |
| **Fly.io** | ❌ No | ✅ Yes | 3 VMs | ⭐⭐ Medium |
| **Railway** | ⚠️ May ask | ✅ Yes | $5/mo | ⭐ Easy |
| **Cyclic** | ❌ No | ❌ Sleeps | Unlimited | ⭐ Easy |

## 🎯 Final Recommendation

**For your rental management app:**

1. **Primary Choice: Render.com**
   - No credit card ✅
   - Easy setup ✅
   - Reliable ✅
   - Perfect for your app size ✅

2. **Alternative: Fly.io**
   - If you want always-on (no sleep)
   - Better performance
   - Slightly more setup

## 🚀 Quick Deploy (Render)

```bash
# 1. Push your code to GitHub
git add .
git commit -m "Ready for deployment"
git push origin main

# 2. Go to render.com
# 3. Connect GitHub → Select repo
# 4. Add environment variables
# 5. Deploy!
```

That's it! Your app will be live in ~10 minutes.

---

## 📝 Environment Variables Checklist

Make sure to set these in your deployment platform:

```env
# Required
DATABASE_URL=postgres://user:pass@host/db?sslmode=require

# Recommended
SERVER_PORT=8080
DB_MAX_CONNECTIONS=10
DB_SSL_MODE=require
LOG_LEVEL=info
```

---

## ❓ Troubleshooting

### Render - Service Won't Start
- Check logs in Render dashboard
- Verify DATABASE_URL is correct
- Ensure Dockerfile is in root directory

### Fly.io - Build Fails
- Run `fly logs` to see errors
- Verify fly.toml syntax
- Check Dockerfile path

### Cold Start Slow
- Normal on free tier
- Use UptimeRobot to keep alive
- Or upgrade to paid plan

---

## 🔗 Helpful Links

- [Render Docs](https://render.com/docs)
- [Fly.io Docs](https://fly.io/docs)
- [Railway Docs](https://docs.railway.app)
- [Neon DB Setup](https://neon.tech/docs)

**Need help?** Check the deployment platform's documentation or support forums.

