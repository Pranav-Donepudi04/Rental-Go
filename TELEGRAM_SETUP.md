# Telegram Bot Setup Guide

## Overview
Telegram notifications use a **Telegram Bot** (not phone numbers). The bot sends messages to specific Telegram chat IDs (unique identifiers for users/chats).

## Step 1: Create a Telegram Bot

1. **Open Telegram** on your phone or desktop
2. **Search for "BotFather"** (official Telegram bot creator)
3. **Start a chat** with BotFather
4. **Send the command**: `/newbot`
5. **Follow the prompts**:
   - Choose a name for your bot (e.g., "Rental Manager Bot")
   - Choose a username (must end with "bot", e.g., "rental_manager_bot")
6. **BotFather will give you a token** that looks like:
   ```
   123456789:ABCdefGHIjklMNOpqrsTUVwxyz
   ```
7. **Save this token** - this is your `TELEGRAM_BOT_TOKEN`

## Step 2: Get Your Chat ID (Owner)

The owner needs to get their Telegram Chat ID. Here are two methods:

### Method 1: Using a Helper Bot (Easiest)

1. **Search for "userinfobot"** on Telegram
2. **Start a chat** with @userinfobot
3. **Send any message** (e.g., "/start")
4. **The bot will reply with your chat ID** (a number like `123456789`)
5. **Save this number** - this is your `TELEGRAM_OWNER_CHAT_ID`

### Method 2: Using Your Bot

1. **Start a chat** with your newly created bot (search for your bot's username)
2. **Send any message** to your bot (e.g., "Hello")
3. **Use this URL** in your browser (replace `YOUR_BOT_TOKEN` with your actual token):
   ```
   https://api.telegram.org/botYOUR_BOT_TOKEN/getUpdates
   ```
4. **Look for the "chat" object** in the JSON response:
   ```json
   {
     "message": {
       "chat": {
         "id": 123456789,
         "first_name": "Your Name"
       }
     }
   }
   ```
5. **The "id" value** is your chat ID

## Step 3: Get Tenant Chat IDs (Optional - for tenant notifications)

For each tenant who wants to receive notifications:

1. **Tenant starts a chat** with your bot
2. **Tenant sends any message** to the bot
3. **Use the same method** as above to get their chat ID
4. **Store tenant chat IDs** in your database (you'll need to add a `telegram_chat_id` field to the tenants table)

## Step 4: Configure Environment Variables

Add these to your `.env` file or environment:

```bash
TELEGRAM_BOT_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz
TELEGRAM_OWNER_CHAT_ID=123456789
```

## Important Notes

### ❌ NOT Phone Numbers
- Telegram chat IDs are **NOT phone numbers**
- They are unique numeric identifiers assigned by Telegram
- Each Telegram user has a unique chat ID

### ✅ How It Works
1. You create a **Telegram Bot** (not a regular Telegram account)
2. Users (owner/tenants) **start a chat** with your bot
3. When they message the bot, Telegram assigns them a **chat ID**
4. Your application uses this **chat ID** to send messages to them

### 🔒 Privacy
- Users must **opt-in** by messaging your bot first
- You can only send messages to users who have started a chat with your bot
- This is a Telegram security feature

## Testing

Once configured, test by:
1. Starting your application
2. The scheduler will run daily at 9 AM
3. Check your Telegram for the reminder messages

## Troubleshooting

### "Chat not found" error
- The user hasn't started a chat with your bot yet
- Ask them to send a message to your bot first

### "Unauthorized" error
- Check that your bot token is correct
- Make sure there are no extra spaces in the token

### Messages not sending
- Verify the chat ID is correct (must be a number)
- Ensure the user has messaged the bot at least once
- Check that your bot token hasn't been revoked

