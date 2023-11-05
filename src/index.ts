import TelegramBot from 'node-telegram-bot-api';
import { db } from './db/database';
import { eq } from 'drizzle-orm';
import { messages } from './db/schema';
import 'dotenv/config';
import { makeExportFile } from './exportProcessor';
import fs from 'node:fs';

const token = process.env.TG_TOKEN as string;
console.log('token', token);
if (!token || !token.length)
  throw new Error('Please provide TG_TOKEN as an ENV variable');

const bot = new TelegramBot(token, { polling: true });

// Got new message, add to db
bot.on('message', async (msg) => {
  const messageId = msg.message_id;
  console.log('Received message', messageId);

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
bot.onText(/\/export/, async (msg, match) => {
  const chatId = msg.chat.id;

  console.log('Exporting chat', chatId);

  const all = await db.query.messages.findMany({
    where: eq(messages.chatId, chatId),
  });

  const exp = makeExportFile(all);

  const tempFileName = `${msg.chat.id}_${new Date().toISOString()}`;
  const tempFilePath = './temp/' + tempFileName + '.md';

  if (!fs.existsSync('./temp')) {
    fs.mkdirSync('./temp');
  }

  fs.writeFileSync(tempFilePath, exp);

  await bot.sendDocument(chatId, tempFilePath);

  fs.unlinkSync(tempFilePath);
});

console.log('Bot is running');
