import { integer, sqliteTable, text } from 'drizzle-orm/sqlite-core';

export const messages = sqliteTable('messages', {
  messageId: integer('messageId').notNull().primaryKey(),
  chatId: integer('chatId').notNull(),
  message: text('message').notNull(),
  date: integer('messageDate').notNull(),
});

export type IMessage = typeof messages.$inferSelect;

export const settings = sqliteTable('users', {
  chatId: integer('chatId').notNull().primaryKey(),
  settings: text('settings', { mode: 'json' }),
});
