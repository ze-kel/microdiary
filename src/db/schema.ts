import { integer, sqliteTable, text } from 'drizzle-orm/sqlite-core';

export const messages = sqliteTable('messages', {
  messageId: integer('messageId').notNull().primaryKey(),
  chatId: integer('chatId').notNull(),
  message: text('message').notNull(),
  date: integer('messageDate').notNull(),
});

const MessageType = messages.$inferSelect;

export type IMessage = typeof MessageType;
