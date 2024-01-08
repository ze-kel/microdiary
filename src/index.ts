import TelegramBot from 'node-telegram-bot-api';
import { db } from './db/database';
import { eq } from 'drizzle-orm';
import { messages } from './db/schema';
import 'dotenv/config';
import { makeExportFile } from './exportProcessor';
import fs from 'node:fs';
import { getSettings, setSetting } from './settings';

const token = process.env.TG_TOKEN as string;
console.log('token', token);
if (!token || !token.length)
  throw new Error('Please provide TG_TOKEN as an ENV variable');

const bot = new TelegramBot(token, { polling: true });

// Got new message, add to db
bot.on('message', async (msg) => {
  const messageId = msg.message_id;

  const message = msg.text;
  if (!message || message.startsWith('/')) return;

  const chatId = msg.chat.id;
  const date = msg.date;

  await db
    .insert(messages)
    .values({ chatId, message, messageId, date })
    .onConflictDoUpdate({ target: messages.messageId, set: { message } });
});

// On edited message update db
bot.on('edited_message', async (msg) => {
  const message = msg.text;
  if (!message || message.startsWith('/')) return;

  const chatId = msg.chat.id;
  const messageId = msg.message_id;
  const date = msg.date;

  await db
    .insert(messages)
    .values({ chatId, message, messageId, date })
    .onConflictDoUpdate({ target: messages.messageId, set: { message } });

  console.log('updated');
});

// Request for export, prepare file and send
bot.onText(/\/export/, async (msg) => {
  const chatId = msg.chat.id;

  console.log('Exporting chat', chatId);

  const all = await db.query.messages.findMany({
    where: eq(messages.chatId, chatId),
  });

  if (!all.length) {
    bot.sendMessage(chatId, 'nothing to export yet');
    return;
  }

  const exp = await makeExportFile(all, chatId);

  const tempFileName = `${msg.chat.id}_${new Date().toISOString()}`;
  const tempFilePath = './temp/' + tempFileName + '.md';

  if (!fs.existsSync('./temp')) {
    fs.mkdirSync('./temp');
  }

  fs.writeFileSync(tempFilePath, exp);

  await bot.sendDocument(chatId, tempFilePath);

  fs.unlinkSync(tempFilePath);
});

bot.onText(/\/settings/, async (msg) => {
  const chatId = msg.chat.id;

  const settings = await getSettings(chatId);

  const text = Object.entries(settings)
    .map(([key, value]) => `${key}: ${value}`)
    .join('\n');

  bot.sendMessage(
    chatId,
    'Current settings:\n' + text + '\nTo change use /set key value'
  );
});

bot.onText(/\/set (.+) (.+)/, async (msg, match) => {
  const chatId = msg.chat.id;

  if (!match) {
    bot.sendMessage(chatId, 'Please provide key and value');
    return;
  }

  try {
    const result = await setSetting({ chatId, key: match[1], value: match[2] });

    bot.sendMessage(chatId, JSON.stringify(result));
  } catch (e) {
    bot.sendMessage(chatId, String(e));
  }
});

bot.onText(/\/wipe/, async (msg, match) => {
  const chatId = msg.chat.id;

  try {
    await db.delete(messages).where(eq(messages.chatId, chatId));
  } catch (e) {
    bot.sendMessage(chatId, String(e));
  }
});
