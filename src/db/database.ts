import { drizzle } from 'drizzle-orm/better-sqlite3';
import { migrate } from 'drizzle-orm/better-sqlite3/migrator';

import Database from 'better-sqlite3';
import * as schema from './schema';

const sqlite = new Database('messages.db');
export const db = drizzle<typeof schema>(sqlite, { schema });

migrate(db, { migrationsFolder: './drizzle' });
