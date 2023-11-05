import { drizzle } from 'drizzle-orm/better-sqlite3';
import { migrate } from 'drizzle-orm/better-sqlite3/migrator';
import fs from 'node:fs';

import Database from 'better-sqlite3';
import * as schema from './schema';

if (!fs.existsSync('./db')) {
  fs.mkdirSync('./db');
}
const sqlite = new Database('./db/messages.db');
export const db = drizzle<typeof schema>(sqlite, { schema });

migrate(db, { migrationsFolder: './drizzle' });
